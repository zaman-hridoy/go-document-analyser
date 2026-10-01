package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/zaman-hridoy/go-document-analyser/internal/config"
)

func main() {
	direction := flag.String(
		"direction",
		"up",
		"migration direction: up or down",
	)

	forceVersion := flag.Int(
		"force",
		-1,
		"for database migration version",
	)

	flag.Parse()

	cfg, err := config.MustLoad()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := sql.Open("postgres", cfg.DatabaseURL)

	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})

	if err != nil {
		log.Fatalf("failed to create migration driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		log.Fatalf("failed to create migration instance: %v", err)
	}

	defer m.Close()

	if *forceVersion >= 0 {
		if err := m.Force(*forceVersion); err != nil {
			log.Fatalf("force migration version: %v", err)
		}

		fmt.Printf("migration version forced to %d\n", *forceVersion)
		return
	}

	switch *direction {

	case "up":
		err = m.Up()

	case "down":
		err = m.Steps(-1)

	default:
		log.Fatalf(
			"invalid migration direction: %s",
			*direction,
		)
	}

	if err != nil && err != migrate.ErrNoChange {
		log.Fatalf("migration failed: %v", err)
	}

	if err == migrate.ErrNoChange {
		log.Println("no migration changes")
		return
	}

	fmt.Printf("migration %s completed successfully\n", *direction)
}
