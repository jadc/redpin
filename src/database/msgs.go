package database

import (
	"context"
	"log"
)

// Demo: https://github.com/practicalgo/go-sqlite-demo/blob/main/app.go#L48

func init() {
	// Connect to database
	db, err := Connect()
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	// Create table
	query := `
        CREATE TABLE IF NOT EXISTS msgs (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            message TEXT NOT NULL
        )
    `
	_, err = db.Instance.ExecContext(context.Background(), query)
	if err != nil {
		log.Fatal("Failed to create table: ", err)
	}
}

func (db *database) AddMessage(message string) error {
    // Connect to database
    db, err := Connect()
    if err != nil {
		log.Fatal("Failed to connect to database: ", err)
    }
    // Insert message
    query := `INSERT INTO msgs (message) VALUES (?)`
    _, err = db.Instance.ExecContext(context.Background(), query, message)
    if err != nil {
        return err
    }
    return nil
}

func (db *database) GetMessages() ([]string, error) {
    // Connect to database
    db, err := Connect()
    if err != nil {
        log.Fatal("Failed to connect to database: ", err)
    }
    // Retrieve messages
    rows, err := db.Instance.QueryContext(
        context.Background(),
        "SELECT message FROM msgs",
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    // Parse messages from SQL rows into string slice
    var messages []string
    for rows.Next() {
        var message string
        err = rows.Scan(&message)
        if err != nil {
            return nil, err
        }
        messages = append(messages, message)
    }
    return messages, nil
}
