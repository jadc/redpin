package database

import (
	"context"
	"fmt"
)

type EmojiQuantity struct {
    EmojiID string
    Count int
}

// createStatsTable creates a pin statistics table for a given guild_id.
func (db *database) createStatsTable(guild_id string) error {
    query := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS stats_%s (
            user_id TEXT NOT NULL,
            emoji_id TEXT NOT NULL,
            received BOOLEAN CHECK (received IN (0, 1)) NOT NULL,
            PRIMARY KEY (user_id)
        )
    `, guild_id)
    _, err = db.Instance.ExecContext(context.Background(), query, guild_id)
    if err != nil {
        return fmt.Errorf("Failed to create stats_%s table: %w", guild_id, err)
    }
    return nil
}

// AddStat inserts a statistic into the guild_id's stats table.
func (db *database) AddStats(guild_id string, user_id string, emoji_id string, received bool) error {
    // Create table if it doesn't exist
    err = db.createStatsTable(guild_id)
    if err != nil {
        return err
    }

    r := 0
    if received {
        r = 1
    }

    // Insert statistic
    query := fmt.Sprintf(`INSERT INTO stats_%s (user_id, emoji_id, received) VALUES (?, ?, ?)`, guild_id)
    _, err = db.Instance.ExecContext(context.Background(), query, user_id, emoji_id, r)
    if err != nil {
        return fmt.Errorf("Failed to insert into table: %w", err)
    }
    return nil
}

// GetStats returns the total number of pins a user has given/received in a guild,
// as well as a detailed breakdown of which specific emojis were used
func (db *database) GetStats(guild_id string, user_id string, received bool) (int, error) {
    // Create guild pins table if it doesn't exist
    err = db.createPinTable(guild_id)
    if err != nil {
        return 0, err
    }

    r := 0
    if received {
        r = 1
    }

    // Query total count
    var count int
    query := fmt.Sprintf(`SELECT COUNT(*) FROM stats_%s WHERE user_id = ? AND received = ?`, guild_id)
    err := db.Instance.QueryRowContext(context.Background(), query, user_id, r).Scan(&count)
    if err != nil {
        return 0, err
    }

    return count, nil
}

// GetDetailedStats returns the total number of pins, and specific emojis used, a user has given/received in a guild,
func (db *database) GetDetailedStats(guild_id string, user_id string, received bool) ([]*EmojiQuantity, error) {
    // Create guild pins table if it doesn't exist
    err = db.createPinTable(guild_id)
    if err != nil {
        return nil, err
    }

    r := 0
    if received {
        r = 1
    }

    // Query count of each emoji
    query := fmt.Sprintf(`SELECT emoji_id, COUNT(*) FROM stats_%s WHERE user_id = ? AND received = ? GROUP BY emoji_id`, guild_id)
    rows, err := db.Instance.QueryContext(context.Background(), query, user_id, r)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    // Convert counts into list of structs
    var quantities []*EmojiQuantity
    for rows.Next() {
        var q EmojiQuantity
        if err := rows.Scan(&q.EmojiID, &q.Count); err != nil {
            return nil, err
        }
        quantities = append(quantities, &q)
    }

    return quantities, nil
}
