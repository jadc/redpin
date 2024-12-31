package database

import (
	"context"
	"fmt"
)

// createWebhookTable creates a guild_id -> webhook_id table.
func (db *database) createWebhookTable() error {
    query := `
        CREATE TABLE IF NOT EXISTS webhooks (
            guild_id TEXT NOT NULL,
            webhook_id TEXT NOT NULL,
            PRIMARY KEY (guild_id, webhook_id)
        )
    `
    _, err = db.Instance.ExecContext(context.Background(), query)
    if err != nil {
        return fmt.Errorf("Failed to create webhooks table: %w", err)
    }
    return nil
}

// AddWebhook inserts a guild_id -> webhook_id pair into the webhooks table.
func (db *database) AddWebhook(guild_id string, webhook_id string) error {
    // Create guild pins table if it doesn't exist
    err = db.createWebhookTable()
    if err != nil {
        return fmt.Errorf("Failed to create table: %w", err)
    }

    // Insert row
    query := "INSERT INTO webhooks (guild_id, webhook_id) VALUES (?, ?)"
    _, err = db.Instance.ExecContext(context.Background(), query, guild_id, webhook_id)
    if err != nil {
        return fmt.Errorf("Failed to insert into table: %w", err)
    }
    return nil
}

// RemoveWebhook deletes all webhooks associated with a guild_id from the webhooks table.
func (db *database) RemoveWebhooks(guild_id string) error {
    // Create guild pins table if it doesn't exist
    err = db.createWebhookTable()
    if err != nil {
        return fmt.Errorf("Failed to create table: %w", err)
    }

    query := "DELETE FROM webhooks WHERE guild_id = ?"
    _, err = db.Instance.ExecContext(context.Background(), query, guild_id)
    if err != nil {
        return fmt.Errorf("Failed to delete from table: %w", err)
    }
    return nil
}

// GetWebhook retrieves the webhook id for a given guild_id.
func (db *database) GetWebhook(guild_id string) (string, error) {
    webhook_id := ""

    // Create guild pins table if it doesn't exist
    err = db.createWebhookTable()
    if err != nil {
        return "", fmt.Errorf("Failed to create table: %w", err)
    }

    // Retrieve pin_id associated with message_id
    err := db.Instance.QueryRowContext(
        context.Background(),
        "SELECT webhook_id FROM webhooks WHERE guild_id = ?", guild_id,
    ).Scan(&webhook_id)

    // Throw up any error
    if err != nil {
        return "", err
    }

    return webhook_id, nil
}
