package migrations

import (
	"database/sql"
	"log"
)

type Migration struct {
	Name string
	Up   func(*sql.DB) error
	Down func(*sql.DB) error
}

var Migrations = []Migration{
	{
		Name: "001_create_connections_table",
		Up: func(db *sql.DB) error {
			_, err := db.Exec(`
			CREATE TABLE IF NOT EXISTS connections (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				client_ip TEXT,
				server_url TEXT,
				timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
			);`)
			return err
		},
		Down: func(db *sql.DB) error {
			_, err := db.Exec(`DROP TABLE IF EXISTS connections;`)
			return err
		},
	},
}

func applyMigration(db *sql.DB, migration Migration) error {
	if _, err := db.Exec("INSERT INTO migrations (name) VALUES (?)", migration.Name); err != nil {
		return err
	}
	if err := migration.Up(db); err != nil {
		return err
	}
	return nil
}

func rollbackMigration(db *sql.DB, migration Migration) error {
	if err := migration.Down(db); err != nil {
		return err
	}
	if _, err := db.Exec("DELETE FROM migrations WHERE name = ?", migration.Name); err != nil {
		return err
	}
	return nil
}

func Migrate(db *sql.DB) {
	var appliedMigrations []string
	rows, err := db.Query("SELECT name FROM migrations ORDER BY applied_at")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			log.Fatal(err)
		}
		appliedMigrations = append(appliedMigrations, name)
	}

	applied := make(map[string]bool)
	for _, name := range appliedMigrations {
		applied[name] = true
	}

	for _, migration := range Migrations {
		if !applied[migration.Name] {
			log.Printf("Applying migration: %s", migration.Name)
			if err := applyMigration(db, migration); err != nil {
				log.Fatalf("Failed to apply migration %s: %v", migration.Name, err)
			}
		}
	}
}

func RollbackLastMigration(db *sql.DB) {
	var lastMigration string
	row := db.QueryRow("SELECT name FROM migrations ORDER BY applied_at DESC LIMIT 1")
	if err := row.Scan(&lastMigration); err != nil {
		if err == sql.ErrNoRows {
			log.Println("No migrations to rollback")
			return
		}
		log.Fatal(err)
	}

	for _, migration := range Migrations {
		if migration.Name == lastMigration {
			log.Printf("Rolling back migration: %s", migration.Name)
			if err := rollbackMigration(db, migration); err != nil {
				log.Fatalf("Failed to rollback migration %s: %v", migration.Name, err)
			}
		}
	}
}
