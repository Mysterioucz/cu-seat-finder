package main

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ponytail: CREATE SCHEMA/TABLE IF NOT EXISTS on boot instead of a migration tool.
// Upgrade path once schema changes get non-trivial: golang-migrate.
func migrate(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, `
		CREATE SCHEMA IF NOT EXISTS library;

		CREATE TABLE IF NOT EXISTS library.libraries (
			id text PRIMARY KEY,
			name_th text NOT NULL,
			name_en text NOT NULL,
			capacity int NOT NULL,
			amenities jsonb NOT NULL DEFAULT '[]',
			image_url text,
			created_at timestamptz NOT NULL DEFAULT now()
		);

		CREATE TABLE IF NOT EXISTS library.floors (
			id text PRIMARY KEY,
			library_id text NOT NULL REFERENCES library.libraries(id),
			name text NOT NULL
		);

		CREATE TABLE IF NOT EXISTS library.zones (
			id text PRIMARY KEY,
			floor_id text NOT NULL REFERENCES library.floors(id),
			name text NOT NULL,
			capacity int NOT NULL
		);

		CREATE TABLE IF NOT EXISTS library.rooms (
			id text PRIMARY KEY,
			library_id text NOT NULL REFERENCES library.libraries(id),
			name text NOT NULL,
			capacity int NOT NULL,
			amenities jsonb NOT NULL DEFAULT '[]'
		);
	`)
	if err != nil {
		return err
	}
	return seed(ctx, db)
}

type seedLibrary struct {
	id, nameTh, nameEn string
	capacity           int
	amenities          []string
	floors             []seedFloor
	rooms              []seedRoom
}
type seedFloor struct {
	id, name string
	zones    []seedZone
}
type seedZone struct {
	id, name string
	capacity int
}
type seedRoom struct {
	id, name  string
	capacity  int
	amenities []string
}

func seedData() []seedLibrary {
	return []seedLibrary{
		{
			id: "lib-central", nameTh: "หอสมุดกลาง", nameEn: "Chula Central Library", capacity: 800,
			amenities: []string{"strong_wifi", "many_outlets", "aircon"},
			floors: []seedFloor{
				{id: "lib-central-f1", name: "Floor 1", zones: []seedZone{{"lib-central-f1-z1", "Silent Zone", 200}}},
				{id: "lib-central-f2", name: "Floor 2", zones: []seedZone{{"lib-central-f2-z1", "Group Study", 150}}},
			},
			rooms: []seedRoom{
				{"room-central-1", "Meeting Room A", 6, []string{"strong_wifi", "whiteboard"}},
				{"room-central-2", "Meeting Room B", 4, []string{"strong_wifi"}},
				{"room-central-3", "Discussion Room C", 8, []string{"strong_wifi", "projector"}},
			},
		},
		{
			id: "lib-engineering", nameTh: "ห้องสมุดคณะวิศวกรรมศาสตร์", nameEn: "Engineering Library", capacity: 400,
			amenities: []string{"strong_wifi", "many_outlets"},
			floors: []seedFloor{
				{id: "lib-eng-f1", name: "Floor 1", zones: []seedZone{{"lib-eng-f1-z1", "Reading Area", 250}}},
			},
			rooms: []seedRoom{
				{"room-eng-1", "Meeting Room A", 6, []string{"strong_wifi", "whiteboard"}},
				{"room-eng-2", "Meeting Room B", 10, []string{"strong_wifi", "projector"}},
			},
		},
		{
			id: "lib-science", nameTh: "ห้องสมุดคณะวิทยาศาสตร์", nameEn: "Science Library", capacity: 350,
			amenities: []string{"weak_wifi", "cold"},
			floors: []seedFloor{
				{id: "lib-sci-f1", name: "Floor 1", zones: []seedZone{{"lib-sci-f1-z1", "Main Hall", 350}}},
			},
			rooms: []seedRoom{
				{"room-sci-1", "Meeting Room A", 4, []string{"weak_wifi"}},
				{"room-sci-2", "Meeting Room B", 6, []string{}},
			},
		},
		{
			id: "lib-arts", nameTh: "ห้องสมุดคณะอักษรศาสตร์", nameEn: "Arts Library", capacity: 300,
			amenities: []string{"quiet", "strong_wifi"},
			floors: []seedFloor{
				{id: "lib-arts-f1", name: "Floor 1", zones: []seedZone{{"lib-arts-f1-z1", "Quiet Zone", 300}}},
			},
			rooms: []seedRoom{
				{"room-arts-1", "Discussion Room A", 5, []string{"strong_wifi"}},
			},
		},
		{
			id: "lib-arch", nameTh: "ห้องสมุดคณะสถาปัตยกรรมศาสตร์", nameEn: "Architecture Library", capacity: 200,
			amenities: []string{"clean", "strong_wifi"},
			floors: []seedFloor{
				{id: "lib-arch-f1", name: "Floor 1", zones: []seedZone{{"lib-arch-f1-z1", "Studio Area", 200}}},
			},
			rooms: []seedRoom{
				{"room-arch-1", "Crit Room A", 12, []string{"projector", "strong_wifi"}},
				{"room-arch-2", "Meeting Room B", 6, []string{}},
			},
		},
		{
			id: "lib-law", nameTh: "ห้องสมุดคณะนิติศาสตร์", nameEn: "Law Library", capacity: 250,
			amenities: []string{"quiet", "many_outlets"},
			floors: []seedFloor{
				{id: "lib-law-f1", name: "Floor 1", zones: []seedZone{{"lib-law-f1-z1", "Main Hall", 250}}},
			},
			rooms: []seedRoom{
				{"room-law-1", "Moot Room A", 8, []string{"strong_wifi", "projector"}},
				{"room-law-2", "Meeting Room B", 4, []string{}},
				{"room-law-3", "Meeting Room C", 4, []string{"strong_wifi"}},
			},
		},
	}
}

func seed(ctx context.Context, db *pgxpool.Pool) error {
	for _, lib := range seedData() {
		amenities, _ := json.Marshal(lib.amenities)
		_, err := db.Exec(ctx, `
			INSERT INTO library.libraries (id, name_th, name_en, capacity, amenities)
			VALUES ($1,$2,$3,$4,$5) ON CONFLICT (id) DO NOTHING`,
			lib.id, lib.nameTh, lib.nameEn, lib.capacity, amenities)
		if err != nil {
			return err
		}
		for _, f := range lib.floors {
			_, err := db.Exec(ctx, `
				INSERT INTO library.floors (id, library_id, name) VALUES ($1,$2,$3)
				ON CONFLICT (id) DO NOTHING`, f.id, lib.id, f.name)
			if err != nil {
				return err
			}
			for _, z := range f.zones {
				_, err := db.Exec(ctx, `
					INSERT INTO library.zones (id, floor_id, name, capacity) VALUES ($1,$2,$3,$4)
					ON CONFLICT (id) DO NOTHING`, z.id, f.id, z.name, z.capacity)
				if err != nil {
					return err
				}
			}
		}
		for _, room := range lib.rooms {
			roomAmenities, _ := json.Marshal(room.amenities)
			_, err := db.Exec(ctx, `
				INSERT INTO library.rooms (id, library_id, name, capacity, amenities) VALUES ($1,$2,$3,$4,$5)
				ON CONFLICT (id) DO NOTHING`, room.id, lib.id, room.name, room.capacity, roomAmenities)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
