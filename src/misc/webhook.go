package misc

import (
	"database/sql"
	"fmt"
    "log"

	"github.com/bwmarrin/discordgo"
	"github.com/jadc/redpin/database"
)

// GetWebhook gets the webhook for a given guild
// Returns the webhook's ID (existing or new) if successful
func GetWebhook(discord *discordgo.Session, guild_id string) (*discordgo.Webhook, error) {
    db, err := database.Connect()
    if err != nil {
        log.Printf("Failed to connect to database: %v", err)
    }

    // Query database for an existing webhook
    webhook_id, err := db.GetWebhook(guild_id)

    // Only throw up error if it's an actual error (not just row not found)
    if err != nil && err != sql.ErrNoRows {
        return nil, fmt.Errorf("Failed to retrieve webhook_id for guild with ID %s: %v", guild_id, err)
    }

    c := db.GetConfig(guild_id)

    // Check that config's pin channel is valid
    _, err = discord.Channel(c.Channel)
    if err != nil {
        return nil, fmt.Errorf("Failed to retrieve channel with ID '%s': %v", c.Channel, err)
    }

    // Retrieve webhook object, create a new one if it's invalid
    webhook, err := discord.Webhook(webhook_id)
    if err != nil {
        log.Printf("Failed to retrieve webhook '%s', creating new one: %v", webhook_id, err)
        webhook, err := createWebhook(discord, guild_id, c.Channel)
        return webhook, err
    }

    // Ensure webhook is pointing to current pin channel
    if webhook.ChannelID != c.Channel {
        log.Printf("Webhook is pointing to wrong channel, updating")

        // Delete incorrect webhook
        err = discord.WebhookDelete(webhook_id)
        if err != nil {
            return nil, fmt.Errorf("Failed to delete outdated webhook '%s': %v", webhook_id, err)
        }

        // Create new webhook in pin channel
        webhook, err := createWebhook(discord, guild_id, c.Channel)
        return webhook, err
    }

    // Return existing webhook if it is valid
    return webhook, nil
}

// CreateWebhook creats a webhook for a given guild to the given pin channel
// Returns the webhook's ID if successful
func createWebhook(discord *discordgo.Session, guild_id string, channel_id string) (*discordgo.Webhook, error) {
    db, err := database.Connect()
    if err != nil {
        log.Printf("Failed to connect to database: %v", err)
    }

    // Create new webhook in given channel
    webhook, err := discord.WebhookCreate(channel_id, "redpin", "")
    if err != nil {
        return nil, fmt.Errorf("Failed to create webhook in channel '%s': %v", channel_id, err)
    }

    err = db.AddWebhook(guild_id, webhook.ID)
    if err != nil {
        return nil, fmt.Errorf("Failed to add webhook to database: %v", err)
    }

    return webhook, nil
}

// GetUserName gets the nickname, or if it is not set, username, for a given user
func getUserName(discord *discordgo.Session, guild_id string, user *discordgo.User) (string, error) {
    member, err := discord.GuildMember(guild_id, user.ID)
    if err != nil {
        return "", fmt.Errorf("Failed to retrieve member with ID '%s': %v", user.ID, err)
    }

    if len(member.Nick) > 0 {
        return member.Nick, nil
    }

    if len(member.User.GlobalName) > 0 {
        return member.User.GlobalName, nil
    }

    return user.Username, nil
}
