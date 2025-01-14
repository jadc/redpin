package database

import (
	"context"
	"fmt"
)

// createWebhookTable creates a guild_id -> webhook pair table.
func (db *database) createWebhookTable() error {
    query := `
        CREATE TABLE IF NOT EXISTS webhooks (
            guild_id TEXT NOT NULL,
            webhook_a TEXT NOT NULL,
            webhook_b TEXT NOT NULL,
            PRIMARY KEY (guild_id)
        )
    `
    _, err = db.Instance.ExecContext(context.Background(), query)
    if err != nil {
        return fmt.Errorf("Failed to create webhooks table: %w", err)
    }
    return nil
}

// SetWebhook creates/updates a guild_id -> webhook pair into the webhooks table.
func (db *database) SetWebhook(guild_id string, webhook_a string, webhook_b string) error {
    // Create guild webhooks table if it doesn't exist
    err := db.createWebhookTable()
    if err != nil {
        return err
    }

    // Create or update row
    query := `
        INSERT INTO webhooks (guild_id, webhook_a, webhook_b) VALUES (?, ?, ?)
        ON CONFLICT (guild_id) DO UPDATE SET webhook_a = ?, webhook_b = ?
    `
    _, err = db.Instance.ExecContext(context.Background(), query,
        guild_id, webhook_a, webhook_b, webhook_a, webhook_b)
    if err != nil {
        return fmt.Errorf("Failed to insert into table: %w", err)
    }

    return nil
}

// GetWebhook retrieves the webhook pair for a given guild_id.
func (db *database) GetWebhook(guild_id string) (string, string, error) {
    // Create guild webhooks table if it doesn't exist
    err := db.createWebhookTable()
    if err != nil {
        return "", "", err
    }

    // Retrieve webhook pair associated with guild_id
    webhook_a, webhook_b := "", ""
    err = db.Instance.QueryRowContext(
        context.Background(),
        "SELECT webhook_a, webhook_b FROM webhooks WHERE guild_id = ?", guild_id,
    ).Scan(&webhook_a, &webhook_b)

    // Throw up any error
    if err != nil || len(webhook_a) == 0 || len(webhook_b) == 0 {
        return "", "", err
    }

    return webhook_a, webhook_b, nil
}
