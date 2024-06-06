package database

import (
	"database/sql"
	"log"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var (
	DB       *sql.DB
	buffer   []ConnectionData
	bufMutex sync.Mutex
	bufSize  = 100
)

type ConnectionData struct {
	ClientIP  string
	ServerURL string
}

func InitDB() {
	var err error
	DB, err = sql.Open("sqlite3", "./loadbalancer.db")
	if err != nil {
		log.Fatal(err)
	}

	sqlStmt := `
    CREATE TABLE IF NOT EXISTS connections (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        client_ip TEXT,
        server_url TEXT,
        timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
    );
	CREATE TABLE IF NOT EXISTS migrations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
    `
	_, err = DB.Exec(sqlStmt)
	if err != nil {
		log.Fatalf("%q: %s\n", err, sqlStmt)
	}

	go processBuffer()
}

func AddConnection(data ConnectionData) {
	bufMutex.Lock()
	defer bufMutex.Unlock()

	buffer = append(buffer, data)
	if len(buffer) >= bufSize {
		writeBufferToDB()
	}
}

func processBuffer() {
	for {
		time.Sleep(10 * time.Second)
		bufMutex.Lock()
		if len(buffer) > 0 {
			writeBufferToDB()
		}
		bufMutex.Unlock()
	}
}

func writeBufferToDB() {
	tx, err := DB.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return
	}

	stmt, err := tx.Prepare("INSERT INTO connections (client_ip, server_url) VALUES (?, ?)")
	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		return
	}
	defer stmt.Close()

	for _, data := range buffer {
		_, err = stmt.Exec(data.ClientIP, data.ServerURL)
		if err != nil {
			log.Printf("Error executing statement: %v", err)
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("Error committing transaction: %v", err)
		return
	}

	buffer = buffer[:0]
}
