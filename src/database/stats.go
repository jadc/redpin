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
            emoji_id TEXT NOT NULL
        )
    `, guild_id)
    _, err = db.Instance.ExecContext(context.Background(), query, guild_id)
    if err != nil {
        return fmt.Errorf("Failed to create stats_%s table: %w", guild_id, err)
    }
    return nil
}

// AddStat inserts a statistic into the guild_id's stats table.
func (db *database) AddStats(guild_id string, user_id string, emoji_id string) error {
    // Create table if it doesn't exist
    err = db.createStatsTable(guild_id)
    if err != nil {
        return err
    }

    // Insert statistic
    query := fmt.Sprintf(`INSERT INTO stats_%s (user_id, emoji_id) VALUES (?, ?)`, guild_id)
    _, err = db.Instance.ExecContext(context.Background(), query, user_id, emoji_id)
    if err != nil {
        return fmt.Errorf("Failed to insert into table: %w", err)
    }
    return nil
}

// GetStats returns the total number of pins a user has received in a guild.
func (db *database) GetStats(guild_id string, user_id string) (int, error) {
    // Create guild pins table if it doesn't exist
    err = db.createPinTable(guild_id)
    if err != nil {
        return 0, err
    }

    // Query total count
    var count int
    query := fmt.Sprintf(`SELECT COUNT(*) FROM stats_%s WHERE user_id = ?`, guild_id)
    err := db.Instance.QueryRowContext(context.Background(), query, user_id).Scan(&count)
    if err != nil {
        return 0, err
    }

    return count, nil
}

// GetDetailedStats returns the total number of pins, with specific emojis used, a user has received in a guild.
func (db *database) GetDetailedStats(guild_id string, user_id string) ([]*EmojiQuantity, error) {
    // Create guild pins table if it doesn't exist
    err = db.createPinTable(guild_id)
    if err != nil {
        return nil, err
    }

    // Query count of each emoji
    query := fmt.Sprintf(`SELECT emoji_id, COUNT(*) FROM stats_%s WHERE user_id = ? GROUP BY emoji_id`, guild_id)
    rows, err := db.Instance.QueryContext(context.Background(), query, user_id)
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
