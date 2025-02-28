package database

import (
	"context"
	"fmt"
    "log"
    "encoding/json"
)

type Config struct {
    Channel     string              `json:"channel"`
    Threshold   int                 `json:"threshold"`
    NSFW        bool                `json:"nsfw"`
    Selfpin     bool                `json:"selfpin"`
    ReplyDepth  int                 `json:"replyDepth"`
    Allowlist   map[string]struct{} `json:"allowlist"`
}

func (c *Config) New() *Config {
    c.Channel = "Set with /redpin channel"
    c.Threshold = 3
    c.NSFW = false
    c.Selfpin = false
    c.ReplyDepth = 1
    c.Allowlist = make(map[string]struct{})
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
func (db *database) GetConfig(guild_id string) *Config {
    // Return config in memory if it is
    if c, ok := db.Config[guild_id]; ok {
        return c
    }

    // Load config from database if it isn't in memory
    // If load fails, use default
    c, err := db.LoadConfig(guild_id)
    if err != nil {
        log.Printf("Using default config for guild '%s': %v\n", guild_id, err)
        db.SaveConfig(guild_id, c)
    }
    db.Config[guild_id] = c
    return c
}

// LoadConfig retrieves the config for a given guild_id.
func (db *database) LoadConfig(guild_id string) (*Config, error) {
    log.Printf("Loading config for guild '%s'\n", guild_id)

    // Create new config object with defaults
    c := new(Config).New()

    // Create guild pins table if it doesn't exist
    if err := db.createConfigTable(); err != nil {
        return c, fmt.Errorf("Failed to create table for guild '%s': %w", guild_id, err)
    }

    // Retrieve config data from database, load default if not found
    var raw string
    err = db.Instance.QueryRow("SELECT json FROM config WHERE guild_id = ?", guild_id).Scan(&raw)
    if err != nil {
        return c, fmt.Errorf("Failed to retrieve config for guild '%s': %w", guild_id, err)
    }

    // Load config into new object
    err := json.Unmarshal([]byte(raw), c)
    if err != nil {
        return c, fmt.Errorf("Failed to unmarshal config for guild '%s': %w", guild_id, err)
    }

    return c, nil
}

// SaveConfig saves to the config for a given guild_id.
func (db *database) SaveConfig(guild_id string, c *Config) error {
    log.Printf("Saving config for guild '%s'\n", guild_id)

    // Create guild pins table if it doesn't exist
    err = db.createConfigTable()
    if err != nil {
        return fmt.Errorf("Failed to create config table: %w", err)
    }

    // Marshal config data into JSON string
    raw, err := json.Marshal(*c)
    if err != nil {
        return fmt.Errorf("Failed to marshal config for guild '%s': %w", guild_id, err)
    }

    // Store JSON string into database row for guild
    _, err = db.Instance.ExecContext(context.Background(), `
        INSERT INTO config (guild_id, json) VALUES (?, ?)
        ON CONFLICT (guild_id) DO UPDATE SET json = ?
    `, guild_id, raw, raw)
    if err != nil {
        return fmt.Errorf("Failed to save config for guild '%s': %w", guild_id, err)
    }

    // Update config in memory
    db.Config[guild_id] = c

    return nil
}
