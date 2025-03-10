package database

import (
    "database/sql"
	"log"
	"os"
    "sync"

    _ "modernc.org/sqlite"
)

// Enclose sqlite3 instance in a struct to have methods attached to it
type database struct {
    Instance *sql.DB
    Config map[string]*Config
}
var db *database

var once sync.Once
var err error

// Connects to database and returns a pointer to it
func Connect() (*database, error) {
    // Only runs once
    once.Do(func(){
        // Retrieve file name from the environment
        database_url := "data.db"
        if x := os.Getenv("DB_FILE"); x != "" {
            log.Println("Using environmental variable 'DB_FILE' for database name.")
            database_url = x
        }

        // Create sqlite connection
        instance, err := sql.Open("sqlite", database_url)
        if err != nil {
            log.Fatal("Failed to create SQLite connection: ", err)
        }

        // Limit writes to one at a time (not ideal)
        instance.SetMaxOpenConns(1)

        // Assign to global variable
        err = instance.Ping()
        db = &database{instance, make(map[string]*Config)}
    })

    // Returns current or new sqlite3 instance (after testing with ping)
    return db, err
}
