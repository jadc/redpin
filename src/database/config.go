package database

import (
	"context"
	"fmt"
    "log"
    "encoding/json"
)

type Config struct {
    Channel     string          `json:"channel"`
    Threshold   int             `json:"threshold"`
    NSFW        bool            `json:"nsfw"`
    Selfpin     bool            `json:"selfpin"`
    Allowlist   map[string]bool `json:"allowlist"`
}

func (c *Config) New() *Config {
    c.Channel = ""
    c.Threshold = 1
    c.NSFW = false
    c.Selfpin = false
    c.Allowlist = make(map[string]bool)
    return c
}

// createConfigTable creates a guild_id -> json data table.
func (db *database) createConfigTable() error {
    query := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS config (
            guild_id TEXT NOT NULL,
            json TEXT NOT NULL,
            PRIMARY KEY (guild_id)
        )
    `)
    _, err = db.Instance.ExecContext(context.Background(), query)
    if err != nil {
        return fmt.Errorf("Failed to create config table: %w", err)
    }

    return nil
}

// GetConfig retrieves the config for a given guild_id, loads if it isn't.
func (db *database) GetConfig(guild_id string) (*Config, error) {
    // Return config in memory if it is
    if c, ok := db.Config[guild_id]; ok {
        return c, nil
    }

    // Load config from database if it isn't in memory
    c, err := db.LoadConfig(guild_id)
    if err != nil {
        return nil, err
    }
    db.Config[guild_id] = c
    return c, err
}

// LoadConfig retrieves the config for a given guild_id.
func (db *database) LoadConfig(guild_id string) (*Config, error) {
    log.Printf("Loading config for guild %s\n", guild_id)
    c := new(Config).New()

    // Create guild pins table if it doesn't exist
    if err := db.createConfigTable(); err != nil {
        return c, fmt.Errorf("Failed to create table for guild %s, using default: %w", guild_id, err)
    }

    // Retrieve config data from database, load default if not found
    var raw string
    err = db.Instance.QueryRow("SELECT json FROM config WHERE guild_id = ?", guild_id).Scan(&raw)
    if err != nil {
        return c, fmt.Errorf("Failed to retrieve config for guild %s, using default: %w", guild_id, err)
    }

    // Update global config map
    err := json.Unmarshal([]byte(raw), c)
    if err != nil {
        return c, fmt.Errorf("Failed to unmarshal config for guild %s, using default: %w", guild_id, err)
    }

    return c, nil
}

// SaveConfig saves to the config for a given guild_id.
func (db *database) SaveConfig(guild_id string, c *Config) error {
    log.Printf("Saving config for guild %s\n", guild_id)

    // Create guild pins table if it doesn't exist
    err = db.createConfigTable()
    if err != nil {
        return fmt.Errorf("Failed to create config table: %w", err)
    }

    // Save config data
    raw, err := json.Marshal(*c)
    if err != nil {
        return fmt.Errorf("Failed to marshal config for guild %s: %w", guild_id, err)
    }

    _, err = db.Instance.ExecContext(context.Background(), `
        INSERT INTO config (guild_id, json) VALUES (?, ?)
        ON CONFLICT (guild_id) DO UPDATE SET json = ?
    `, guild_id, raw, raw)
    if err != nil {
        return fmt.Errorf("Failed to save config for guild %s: %w", guild_id, err)
    }

    return nil
}
