package database

import (
	"context"
	"fmt"
)

// createPinTable creates a message_id -> pin_id table for a given guild_id.
func (db *database) createPinTable(guild_id string) error {
    query := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS pins_%s (
            message_id TEXT NOT NULL,
            pin_channel_id TEXT NOT NULL,
            pin_id TEXT NOT NULL,
            PRIMARY KEY (message_id, pin_channel_id, pin_id)
        )
    `, guild_id)
    _, err = db.Instance.ExecContext(context.Background(), query, guild_id)
    if err != nil {
        return fmt.Errorf("Failed to create pins_%s table: %w", guild_id, err)
    }
    return nil
}

// AddPin inserts a message_id -> pin_id pair into the guild_id table.
func (db *database) AddPin(guild_id string, pin_channel_id string, message_id string, pin_id string) error {
    // Create guild pins table if it doesn't exist
    err = db.createPinTable(guild_id)
    if err != nil {
        return fmt.Errorf("Failed to create table: %w", err)
    }

    // Insert message
    query := fmt.Sprintf(`INSERT INTO pins_%s (message_id, pin_channel_id, pin_id) VALUES (?, ?, ?)`, guild_id)
    _, err = db.Instance.ExecContext(context.Background(), query, message_id, pin_channel_id, pin_id)
    if err != nil {
        return fmt.Errorf("Failed to insert into table: %w", err)
    }
    return nil
}

// GetPin retrieves the pin message id from the guild_id table given a guild_id and message_id.
func (db *database) GetPin(guild_id string, message_id string) (string, string, error) {
    // Create guild pins table if it doesn't exist
    err = db.createPinTable(guild_id)
    if err != nil {
        return "", "", fmt.Errorf("Failed to create table: %w", err)
    }

    // Retrieve pin_id associated with message_id
    pin_id := ""
    pin_channel_id := ""
    err := db.Instance.QueryRowContext(
        context.Background(),
        "SELECT pin_channel_id, pin_id FROM pins_" + guild_id + " WHERE message_id = ?", message_id,
    ).Scan(&pin_channel_id, &pin_id)

    // Throw up any error
    if err != nil {
        return "", "", err
    }

    return pin_channel_id, pin_id, nil
}
