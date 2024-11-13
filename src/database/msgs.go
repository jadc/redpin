package database

import (
	"context"
	"fmt"
    "log"
)

// CreateTable creates a message_id -> pin_id table for a given guild_id.
func (db *database) CreateTable(guild_id string) error {
    query := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS pins_%s (
            message_id TEXT NOT NULL,
            pin_id TEXT NOT NULL,
            PRIMARY KEY (message_id, pin_id)
        )
    `, guild_id)
    _, err = db.Instance.ExecContext(context.Background(), query, guild_id)
    if err != nil {
        return fmt.Errorf("Failed to create pins_%s table: %w", guild_id, err)
    }
    return nil
}

// AddPin inserts a message_id -> pin_id pair into the guild_id table.
func (db *database) AddPin(guild_id string, message_id string, pin_id string) error {
    log.Printf("Adding %s to pins_%s table\n", message_id, guild_id)

    // Create guild pins table if it doesn't exist
    err = db.CreateTable(guild_id)
    if err != nil {
        return fmt.Errorf("Failed to create table: %w", err)
    }

    // Insert message
    query := fmt.Sprintf(`INSERT INTO pins_%s (message_id, pin_id) VALUES (?, ?)`, guild_id)
    _, err = db.Instance.ExecContext(context.Background(), query, message_id, pin_id)
    if err != nil {
        return fmt.Errorf("Failed to insert into table: %w", err)
    }
    return nil
}

func (db *database) GetMessages(guild_id string) ([]string, error) {
    log.Printf("Retrieving messages from pins_%s table\n", guild_id)

    // Create guild pins table if it doesn't exist
    err = db.CreateTable(guild_id)
    if err != nil {
        return nil, err
    }

    // Retrieve messages
    rows, err := db.Instance.QueryContext(
        context.Background(),
        "SELECT * FROM pins_" + guild_id,
    )
    if err != nil {
        return nil, fmt.Errorf("Failed to retrieve pinned messages: %w", err)
    }
    defer rows.Close()

    // Parse messages from SQL rows into string slice
    var messages []string
    for rows.Next() {
        var message_id string
        var pin_id string

        err = rows.Scan(&message_id, &pin_id)
        log.Printf("Found message %s -> %s\n", message_id, pin_id)
        if err != nil {
            return nil, err
        }
        messages = append(messages, message_id + "-" + pin_id)
    }
    return messages, nil
}
