package database

import (
	"context"
	"fmt"
)

type UserStats struct {
    UserID string
    Total int
    Emojis []*EmojiQuantity
}

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

// GetStats returns the total number of pins, with specific emojis used, a user has received in a guild.
func (db *database) GetStats(guild_id string, user_id string) (*UserStats, error) {
    // Create guild pins table if it doesn't exist
    err = db.createStatsTable(guild_id)
    if err != nil {
        return nil, err
    }

    stats := &UserStats{ UserID: user_id }

    // Query total count
    var count int
    query := fmt.Sprintf(`
        SELECT COUNT(*)
        FROM stats_%s
        WHERE user_id = ?`, guild_id)
    err := db.Instance.QueryRowContext(context.Background(), query, user_id).Scan(&count)
    if err != nil {
        return nil, err
    }
    stats.Total = count

    // Query count of each emoji
    query = fmt.Sprintf(`
        SELECT emoji_id, COUNT(*) as count
        FROM stats_%s
        WHERE user_id = ?
        GROUP BY emoji_id
        ORDER BY count DESC`, guild_id)
    rows, err := db.Instance.QueryContext(context.Background(), query, user_id)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    // Convert counts into list of structs
    for rows.Next() {
        var q EmojiQuantity
        if err := rows.Scan(&q.EmojiID, &q.Count); err != nil {
            return nil, err
        }
        stats.Emojis = append(stats.Emojis, &q)
    }

    return stats, nil
}

// GetLeaderboard returns the top ten users in a guild with the most total number of pins, with specific emojis used.
func (db *database) GetLeaderboard(guild_id string) ([]*UserStats, error) {
    // Create guild pins table if it doesn't exist
    err = db.createStatsTable(guild_id)
    if err != nil {
        return nil, err
    }

    // Get top ten users
    query := fmt.Sprintf(`
        SELECT user_id, COUNT(*) as count
        FROM stats_%s
        GROUP BY user_id
        ORDER BY count DESC
        LIMIT 10`, guild_id)
    rows, err := db.Instance.QueryContext(context.Background(), query)
    if err != nil {
        return nil, err
    }

    // Get top 10 user IDs
    var users []string
    for rows.Next() {
        var user_id string
        var count int
        if err := rows.Scan(&user_id, &count); err != nil {
            return nil, err
        }
        users = append(users, user_id)
    }
    rows.Close()

    // Get detailed stats for each user
    stats_list := []*UserStats{}
    for _, user_id := range users {
        stats, err := db.GetStats(guild_id, user_id);
        if err != nil {
            return nil, err
        }
        stats_list = append(stats_list, stats)
    }

    return stats_list, nil
}
