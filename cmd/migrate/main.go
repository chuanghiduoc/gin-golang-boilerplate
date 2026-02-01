package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"backend-gin/internal/infrastructure/config"
)

func main() {
	var direction string
	var steps int

	flag.StringVar(&direction, "direction", "up", "Migration direction: up, down, or force")
	flag.IntVar(&steps, "steps", 0, "Number of migrations to run (0 = all)")
	flag.Parse()

	cfg := config.Load()
	dsn := cfg.Database.DSN()

	m, err := migrate.New("file://db/migrations", dsn)
	if err != nil {
		log.Fatalf("Failed to create migrate instance: %v", err)
	}
	defer m.Close()

	switch direction {
	case "up":
		if steps > 0 {
			err = m.Steps(steps)
		} else {
			err = m.Up()
		}
	case "down":
		if steps > 0 {
			err = m.Steps(-steps)
		} else {
			err = m.Down()
		}
	case "force":
		version := steps
		err = m.Force(version)
	default:
		log.Fatalf("Unknown direction: %s", direction)
	}

	if err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Migration failed: %v", err)
	}

	version, dirty, _ := m.Version()
	fmt.Printf("Migration completed. Current version: %d, Dirty: %v\n", version, dirty)
	os.Exit(0)
}
