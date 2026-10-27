package main

import (
	"context"
	"io"
	"net/http"

	pb "github.com/cu-seat-finder/backend/gen/library/v1"
	"github.com/cu-seat-finder/backend/internal/httpx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// libraryRoutes is the REST facade over LibraryService (moved here from the
// Next.js route handlers, ADR-008). JSON shapes follow API-CONTRACT.md.
func (g *gateway) libraryRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/libraries", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		g.call(w, r, http.StatusOK, func(ctx context.Context, o grpc.CallOption) (any, error) {
			res, err := g.library.ListLibraries(ctx, &pb.ListLibrariesRequest{Q: q.Get("q"), Amenity: q.Get("amenity")}, o)
			if err != nil {
				return nil, err
			}
			out := make([]libraryJSON, 0, len(res.Libraries))
			for _, l := range res.Libraries {
				out = append(out, toLibrary(l))
			}
			return out, nil
		})
	})
	mux.HandleFunc("GET /api/libraries/{id}", func(w http.ResponseWriter, r *http.Request) {
		g.call(w, r, http.StatusOK, func(ctx context.Context, o grpc.CallOption) (any, error) {
			l, err := g.library.GetLibrary(ctx, &pb.GetLibraryRequest{Id: r.PathValue("id")}, o)
			return libraryOrNil(l), err
		})
	})
	mux.HandleFunc("GET /api/libraries/{id}/rooms", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		g.call(w, r, http.StatusOK, func(ctx context.Context, o grpc.CallOption) (any, error) {
			// ListRooms treats an empty library_id as "all rooms", so check the library exists first (404).
			if _, err := g.library.GetLibrary(ctx, &pb.GetLibraryRequest{Id: id}, o); err != nil {
				return nil, err
			}
			res, err := g.library.ListRooms(ctx, &pb.ListRoomsRequest{LibraryId: id}, o)
			if err != nil {
				return nil, err
			}
			out := make([]roomJSON, 0, len(res.Rooms))
			for _, rm := range res.Rooms {
				out = append(out, toRoom(rm))
			}
			return out, nil
		})
	})
	mux.HandleFunc("POST /api/libraries", func(w http.ResponseWriter, r *http.Request) {
		l := &pb.Library{}
		if !decode(w, r, l) {
			return
		}
		g.call(w, r, http.StatusCreated, func(ctx context.Context, o grpc.CallOption) (any, error) {
			res, err := g.library.CreateLibrary(ctx, &pb.CreateLibraryRequest{Library: l}, o)
			return libraryOrNil(res), err
		})
	})
	mux.HandleFunc("PUT /api/libraries/{id}", func(w http.ResponseWriter, r *http.Request) {
		l := &pb.Library{}
		if !decode(w, r, l) {
			return
		}
		l.Id = r.PathValue("id")
		g.call(w, r, http.StatusOK, func(ctx context.Context, o grpc.CallOption) (any, error) {
			res, err := g.library.UpdateLibrary(ctx, &pb.UpdateLibraryRequest{Library: l}, o)
			return libraryOrNil(res), err
		})
	})
	mux.HandleFunc("DELETE /api/libraries/{id}", func(w http.ResponseWriter, r *http.Request) {
		g.call(w, r, http.StatusNoContent, func(ctx context.Context, o grpc.CallOption) (any, error) {
			_, err := g.library.DeleteLibrary(ctx, &pb.DeleteLibraryRequest{Id: r.PathValue("id")}, o)
			return nil, err
		})
	})

	mux.HandleFunc("GET /api/rooms/{id}", func(w http.ResponseWriter, r *http.Request) {
		g.call(w, r, http.StatusOK, func(ctx context.Context, o grpc.CallOption) (any, error) {
			rm, err := g.library.GetRoom(ctx, &pb.GetRoomRequest{Id: r.PathValue("id")}, o)
			return roomOrNil(rm), err
		})
	})
	mux.HandleFunc("POST /api/rooms", func(w http.ResponseWriter, r *http.Request) {
		rm := &pb.Room{}
		if !decode(w, r, rm) {
			return
		}
		g.call(w, r, http.StatusCreated, func(ctx context.Context, o grpc.CallOption) (any, error) {
			res, err := g.library.CreateRoom(ctx, &pb.CreateRoomRequest{Room: rm}, o)
			return roomOrNil(res), err
		})
	})
	mux.HandleFunc("PUT /api/rooms/{id}", func(w http.ResponseWriter, r *http.Request) {
		rm := &pb.Room{}
		if !decode(w, r, rm) {
			return
		}
		rm.Id = r.PathValue("id")
		g.call(w, r, http.StatusOK, func(ctx context.Context, o grpc.CallOption) (any, error) {
			res, err := g.library.UpdateRoom(ctx, &pb.UpdateRoomRequest{Room: rm}, o)
			return roomOrNil(res), err
		})
	})
	mux.HandleFunc("DELETE /api/rooms/{id}", func(w http.ResponseWriter, r *http.Request) {
		g.call(w, r, http.StatusNoContent, func(ctx context.Context, o grpc.CallOption) (any, error) {
			_, err := g.library.DeleteRoom(ctx, &pb.DeleteRoomRequest{Id: r.PathValue("id")}, o)
			return nil, err
		})
	})
}

// call runs one or more Library RPCs with the caller's token as gRPC metadata
// and the gateway's upstream timeout, then writes the JSON result (or the
// mapped error). The CallOption records which instance answered, for the log.
func (g *gateway) call(w http.ResponseWriter, r *http.Request, ok int, fn func(context.Context, grpc.CallOption) (any, error)) {
	ctx, cancel := context.WithTimeout(r.Context(), g.timeout)
	defer cancel()
	if auth := r.Header.Get("Authorization"); auth != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", auth)
	}
	var p peer.Peer
	out, err := fn(ctx, grpc.Peer(&p))
	if p.Addr != nil {
		setUpstream(r.Context(), p.Addr.String())
	}
	w.Header().Set("Cache-Control", "no-store")
	if err != nil {
		st := status.Convert(err)
		httpx.WriteError(w, grpcToHTTP(st.Code()), grpcCodeName(st.Code()), st.Message())
		return
	}
	if ok == http.StatusNoContent {
		w.WriteHeader(ok)
		return
	}
	httpx.WriteJSON(w, ok, out)
}

func decode(w http.ResponseWriter, r *http.Request, m proto.Message) bool {
	body, err := io.ReadAll(r.Body)
	if err == nil {
		err = protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(body, m)
	}
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_argument", "invalid JSON body")
		return false
	}
	return true
}

// grpcToHTTP follows the status mapping in API-CONTRACT.md.
func grpcToHTTP(c codes.Code) int {
	switch c {
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.NotFound:
		return http.StatusNotFound
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.FailedPrecondition:
		return http.StatusConflict
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	}
	return http.StatusInternalServerError
}

var codeNames = map[codes.Code]string{
	codes.InvalidArgument:    "invalid_argument",
	codes.NotFound:           "not_found",
	codes.PermissionDenied:   "permission_denied",
	codes.FailedPrecondition: "failed_precondition",
	codes.Unavailable:        "unavailable",
	codes.Unauthenticated:    "unauthenticated",
	codes.DeadlineExceeded:   "deadline_exceeded",
}

func grpcCodeName(c codes.Code) string {
	if n, ok := codeNames[c]; ok {
		return n
	}
	return "internal"
}

// JSON shapes from API-CONTRACT.md: snake_case, lists never null, image_url omitted when empty.
type zoneJSON struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Capacity int32  `json:"capacity"`
}

type floorJSON struct {
	ID    string     `json:"id"`
	Name  string     `json:"name"`
	Zones []zoneJSON `json:"zones"`
}

type libraryJSON struct {
	ID        string      `json:"id"`
	NameTh    string      `json:"name_th"`
	NameEn    string      `json:"name_en"`
	Capacity  int32       `json:"capacity"`
	Amenities []string    `json:"amenities"`
	Floors    []floorJSON `json:"floors"`
	ImageURL  string      `json:"image_url,omitempty"`
}

type roomJSON struct {
	ID        string   `json:"id"`
	LibraryID string   `json:"library_id"`
	Name      string   `json:"name"`
	Capacity  int32    `json:"capacity"`
	Amenities []string `json:"amenities"`
}

func toLibrary(l *pb.Library) libraryJSON {
	out := libraryJSON{
		ID: l.Id, NameTh: l.NameTh, NameEn: l.NameEn, Capacity: l.Capacity,
		Amenities: nonNil(l.Amenities), Floors: []floorJSON{}, ImageURL: l.ImageUrl,
	}
	for _, f := range l.Floors {
		fj := floorJSON{ID: f.Id, Name: f.Name, Zones: []zoneJSON{}}
		for _, z := range f.Zones {
			fj.Zones = append(fj.Zones, zoneJSON{ID: z.Id, Name: z.Name, Capacity: z.Capacity})
		}
		out.Floors = append(out.Floors, fj)
	}
	return out
}

func toRoom(r *pb.Room) roomJSON {
	return roomJSON{ID: r.Id, LibraryID: r.LibraryId, Name: r.Name, Capacity: r.Capacity, Amenities: nonNil(r.Amenities)}
}

func libraryOrNil(l *pb.Library) any {
	if l == nil {
		return nil
	}
	return toLibrary(l)
}

func roomOrNil(r *pb.Room) any {
	if r == nil {
		return nil
	}
	return toRoom(r)
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
