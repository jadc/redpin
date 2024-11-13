package database

import (
    "database/sql"
	"log"
	"os"
    "sync"

    _ "modernc.org/sqlite"
)

// Enclose sqlite3 instance in a struct to have methods attached to it
type database struct { Instance *sql.DB }
var db *database

var once sync.Once
var err error

// Connects to database and returns a pointer to it
func Connect() (*database, error) {
    // Only runs once
    once.Do(func(){
        // Retrieve file name from the environment
        database_url := "data.db"
        if x := os.Getenv("SQLITE_FILE_NAME"); x != "" {
            log.Println("Using environmental variable 'SQLITE_FILE_NAME' for database name.")
            database_url = x
        }

        // Create sqlite connection
        instance, err := sql.Open("sqlite", database_url)
        if err != nil {
            log.Fatal("Failed to create SQLite connection: ", err)
        }

        // Assign to global variable
        err = instance.Ping()
        db = &database{instance}
    })

    // Returns current or new sqlite3 instance (after testing with ping)
    return db, err
}
