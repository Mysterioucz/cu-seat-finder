package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	pb "github.com/cu-seat-finder/backend/gen/library/v1"
	"github.com/cu-seat-finder/backend/internal/idgen"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	pb.UnimplementedLibraryServiceServer
	db *pgxpool.Pool
}

var errInternal = status.Error(codes.Internal, "internal error")

func bad(msg string) error { return status.Error(codes.InvalidArgument, msg) }

// fkViolation reports a Postgres foreign_key_violation (23503).
func fkViolation(err error) bool {
	var pe *pgconn.PgError
	return errors.As(err, &pe) && pe.Code == "23503"
}

func jsonList(l []string) []byte {
	if l == nil {
		l = []string{}
	}
	b, _ := json.Marshal(l)
	return b
}

// ---- libraries ----

func (s *server) ListLibraries(ctx context.Context, req *pb.ListLibrariesRequest) (*pb.ListLibrariesResponse, error) {
	q := strings.ToLower(req.Q)
	rows, err := s.db.Query(ctx, "SELECT id FROM library.libraries ORDER BY id")
	if err != nil {
		return nil, errInternal
	}
	ids, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, errInternal
	}
	out := &pb.ListLibrariesResponse{Libraries: []*pb.Library{}}
	for _, id := range ids {
		lib, err := loadLibrary(ctx, s.db, id)
		if err != nil {
			continue
		}
		if q != "" && !strings.Contains(strings.ToLower(lib.NameEn), q) && !strings.Contains(strings.ToLower(lib.NameTh), q) {
			continue
		}
		if req.Amenity != "" && !contains(lib.Amenities, req.Amenity) {
			continue
		}
		out.Libraries = append(out.Libraries, lib)
	}
	return out, nil
}

func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

func (s *server) GetLibrary(ctx context.Context, req *pb.GetLibraryRequest) (*pb.Library, error) {
	lib, err := loadLibrary(ctx, s.db, req.Id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, status.Error(codes.NotFound, "library not found")
	}
	if err != nil {
		return nil, errInternal
	}
	return lib, nil
}

func validLibrary(l *pb.Library) error {
	if l == nil {
		return bad("library is required")
	}
	if strings.TrimSpace(l.NameTh) == "" || strings.TrimSpace(l.NameEn) == "" {
		return bad("name_th and name_en are required")
	}
	if l.Capacity <= 0 {
		return bad("capacity must be > 0")
	}
	for _, f := range l.Floors {
		if strings.TrimSpace(f.Name) == "" {
			return bad("floor name is required")
		}
		for _, z := range f.Zones {
			if strings.TrimSpace(z.Name) == "" || z.Capacity <= 0 {
				return bad("zone needs a name and capacity > 0")
			}
		}
	}
	return nil
}

// insertFloors writes floors/zones with deterministic, ordered ids (<lib>-fNN, <floor>-zNN)
// so list order survives the ORDER BY id in loadLibrary.
func insertFloors(ctx context.Context, tx pgx.Tx, libID string, floors []*pb.Floor) error {
	for i, f := range floors {
		fid := fmt.Sprintf("%s-f%02d", libID, i+1)
		if _, err := tx.Exec(ctx, "INSERT INTO library.floors (id, library_id, name) VALUES ($1,$2,$3)", fid, libID, f.Name); err != nil {
			return err
		}
		for j, z := range f.Zones {
			if _, err := tx.Exec(ctx, "INSERT INTO library.zones (id, floor_id, name, capacity) VALUES ($1,$2,$3,$4)",
				fmt.Sprintf("%s-z%02d", fid, j+1), fid, z.Name, z.Capacity); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *server) CreateLibrary(ctx context.Context, req *pb.CreateLibraryRequest) (*pb.Library, error) {
	l := req.Library
	if err := validLibrary(l); err != nil {
		return nil, err
	}
	id := idgen.New()
	err := pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `INSERT INTO library.libraries (id, name_th, name_en, capacity, amenities, image_url)
			VALUES ($1,$2,$3,$4,$5,NULLIF($6,''))`, id, l.NameTh, l.NameEn, l.Capacity, jsonList(l.Amenities), l.ImageUrl); err != nil {
			return err
		}
		return insertFloors(ctx, tx, id, l.Floors)
	})
	if err != nil {
		return nil, errInternal
	}
	return loadLibrary(ctx, s.db, id)
}

func (s *server) UpdateLibrary(ctx context.Context, req *pb.UpdateLibraryRequest) (*pb.Library, error) {
	l := req.Library
	if err := validLibrary(l); err != nil {
		return nil, err
	}
	if l.Id == "" {
		return nil, bad("library.id is required")
	}
	err := pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `UPDATE library.libraries SET name_th=$2, name_en=$3, capacity=$4, amenities=$5, image_url=NULLIF($6,'')
			WHERE id=$1`, l.Id, l.NameTh, l.NameEn, l.Capacity, jsonList(l.Amenities), l.ImageUrl)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return pgx.ErrNoRows
		}
		if err := deleteFloors(ctx, tx, l.Id); err != nil {
			return err
		}
		return insertFloors(ctx, tx, l.Id, l.Floors)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, status.Error(codes.NotFound, "library not found")
	}
	if err != nil {
		return nil, errInternal
	}
	return loadLibrary(ctx, s.db, l.Id)
}

func deleteFloors(ctx context.Context, tx pgx.Tx, libID string) error {
	if _, err := tx.Exec(ctx, "DELETE FROM library.zones WHERE floor_id IN (SELECT id FROM library.floors WHERE library_id=$1)", libID); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, "DELETE FROM library.floors WHERE library_id=$1", libID)
	return err
}

func (s *server) DeleteLibrary(ctx context.Context, req *pb.DeleteLibraryRequest) (*pb.DeleteResponse, error) {
	err := pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
		if err := deleteFloors(ctx, tx, req.Id); err != nil {
			return err
		}
		tag, err := tx.Exec(ctx, "DELETE FROM library.libraries WHERE id=$1", req.Id)
		if err == nil && tag.RowsAffected() == 0 {
			return pgx.ErrNoRows
		}
		return err
	})
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return nil, status.Error(codes.NotFound, "library not found")
	case fkViolation(err): // rooms still reference it
		return nil, status.Error(codes.FailedPrecondition, "library still has rooms")
	case err != nil:
		return nil, errInternal
	}
	return &pb.DeleteResponse{}, nil
}

// ---- rooms ----

func (s *server) ListRooms(ctx context.Context, req *pb.ListRoomsRequest) (*pb.ListRoomsResponse, error) {
	var rooms []*pb.Room
	var err error
	if req.LibraryId == "" {
		rooms, err = loadRooms(ctx, s.db, "")
	} else {
		rooms, err = loadRooms(ctx, s.db, "WHERE library_id=$1", req.LibraryId)
	}
	if err != nil {
		return nil, errInternal
	}
	return &pb.ListRoomsResponse{Rooms: rooms}, nil
}

func (s *server) GetRoom(ctx context.Context, req *pb.GetRoomRequest) (*pb.Room, error) {
	rooms, err := loadRooms(ctx, s.db, "WHERE id=$1", req.Id)
	if err != nil {
		return nil, errInternal
	}
	if len(rooms) == 0 {
		return nil, status.Error(codes.NotFound, "room not found")
	}
	return rooms[0], nil
}

func validRoom(r *pb.Room) error {
	if r == nil {
		return bad("room is required")
	}
	if r.LibraryId == "" {
		return bad("room.library_id is required")
	}
	if strings.TrimSpace(r.Name) == "" {
		return bad("room name is required")
	}
	if r.Capacity <= 0 {
		return bad("capacity must be > 0")
	}
	return nil
}

func (s *server) CreateRoom(ctx context.Context, req *pb.CreateRoomRequest) (*pb.Room, error) {
	r := req.Room
	if err := validRoom(r); err != nil {
		return nil, err
	}
	id := idgen.New()
	_, err := s.db.Exec(ctx, "INSERT INTO library.rooms (id, library_id, name, capacity, amenities) VALUES ($1,$2,$3,$4,$5)",
		id, r.LibraryId, r.Name, r.Capacity, jsonList(r.Amenities))
	if fkViolation(err) {
		return nil, bad("library_id does not exist")
	}
	if err != nil {
		return nil, errInternal
	}
	return s.GetRoom(ctx, &pb.GetRoomRequest{Id: id})
}

func (s *server) UpdateRoom(ctx context.Context, req *pb.UpdateRoomRequest) (*pb.Room, error) {
	r := req.Room
	if err := validRoom(r); err != nil {
		return nil, err
	}
	if r.Id == "" {
		return nil, bad("room.id is required")
	}
	tag, err := s.db.Exec(ctx, "UPDATE library.rooms SET library_id=$2, name=$3, capacity=$4, amenities=$5 WHERE id=$1",
		r.Id, r.LibraryId, r.Name, r.Capacity, jsonList(r.Amenities))
	if fkViolation(err) {
		return nil, bad("library_id does not exist")
	}
	if err != nil {
		return nil, errInternal
	}
	if tag.RowsAffected() == 0 {
		return nil, status.Error(codes.NotFound, "room not found")
	}
	return s.GetRoom(ctx, &pb.GetRoomRequest{Id: r.Id})
}

func (s *server) DeleteRoom(ctx context.Context, req *pb.DeleteRoomRequest) (*pb.DeleteResponse, error) {
	tag, err := s.db.Exec(ctx, "DELETE FROM library.rooms WHERE id=$1", req.Id)
	if err != nil {
		return nil, errInternal
	}
	if tag.RowsAffected() == 0 {
		return nil, status.Error(codes.NotFound, "room not found")
	}
	return &pb.DeleteResponse{}, nil
}

// ---- loaders ----

func loadLibrary(ctx context.Context, db *pgxpool.Pool, id string) (*pb.Library, error) {
	lib := &pb.Library{Amenities: []string{}, Floors: []*pb.Floor{}}
	var amenities []byte
	var img *string
	row := db.QueryRow(ctx, "SELECT id, name_th, name_en, capacity, amenities, image_url FROM library.libraries WHERE id=$1", id)
	if err := row.Scan(&lib.Id, &lib.NameTh, &lib.NameEn, &lib.Capacity, &amenities, &img); err != nil {
		return nil, err
	}
	_ = json.Unmarshal(amenities, &lib.Amenities)
	if img != nil {
		lib.ImageUrl = *img
	}

	frows, err := db.Query(ctx, "SELECT id, name FROM library.floors WHERE library_id=$1 ORDER BY id", id)
	if err != nil {
		return nil, err
	}
	defer frows.Close()
	for frows.Next() {
		f := &pb.Floor{Zones: []*pb.Zone{}}
		if err := frows.Scan(&f.Id, &f.Name); err != nil {
			return nil, err
		}
		lib.Floors = append(lib.Floors, f)
	}
	frows.Close()

	for _, f := range lib.Floors {
		zrows, err := db.Query(ctx, "SELECT id, name, capacity FROM library.zones WHERE floor_id=$1 ORDER BY id", f.Id)
		if err != nil {
			return nil, err
		}
		for zrows.Next() {
			z := &pb.Zone{}
			if err := zrows.Scan(&z.Id, &z.Name, &z.Capacity); err != nil {
				zrows.Close()
				return nil, err
			}
			f.Zones = append(f.Zones, z)
		}
		zrows.Close()
	}
	return lib, nil
}

func loadRooms(ctx context.Context, db *pgxpool.Pool, where string, args ...any) ([]*pb.Room, error) {
	rows, err := db.Query(ctx, "SELECT id, library_id, name, capacity, amenities FROM library.rooms "+where+" ORDER BY id", args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rooms := []*pb.Room{}
	for rows.Next() {
		room := &pb.Room{Amenities: []string{}}
		var amenities []byte
		if err := rows.Scan(&room.Id, &room.LibraryId, &room.Name, &room.Capacity, &amenities); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(amenities, &room.Amenities)
		rooms = append(rooms, room)
	}
	return rooms, rows.Err()
}
