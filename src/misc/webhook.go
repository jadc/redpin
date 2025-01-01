package misc

import (
	"database/sql"
	"fmt"
    "log"
    "net/http"
    "time"

	"github.com/bwmarrin/discordgo"
	"github.com/jadc/redpin/database"
)

// Hashmap of webhook id -> last used timestamp
// Used to prevent many pins in quick succession from being merged into one message
// Technically, if the bot is restarted, timestamps are reset and
// messages may get merged, but this is a very unlikely scenario
var timestamps = make(map[string]int64)
var MERGE_TIME int64 = 10*60*1000

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

    // Discard current webhook if it is too young
    if timestamp, ok := timestamps[webhook_id]; ok {
        if time.Now().UnixMilli() - timestamp < MERGE_TIME {
            log.Printf("Current webhook '%s' is too young, creating a new one", webhook_id)

            // Clear outdated webhooks
            err = discord.WebhookDelete(webhook_id)
            if err != nil {
                return nil, fmt.Errorf("Failed to delete outdated webhook '%s': %v", webhook_id, err)
            }
            delete(timestamps, webhook_id)

            return createWebhook(discord, guild_id, c.Channel)
        }
    }

    // Update timestamps
    timestamps[webhook_id] = time.Now().UnixMilli()

    // Retrieve webhook object, create a new one if it's invalid
    webhook, err := discord.Webhook(webhook_id)
    if err != nil {
        log.Printf("Failed to retrieve webhook '%s', creating a new one: %v", webhook_id, err)
        return createWebhook(discord, guild_id, c.Channel)
    }

    // Ensure webhook is pointing to current pin channel
    if webhook.ChannelID != c.Channel {
        log.Printf("Webhook is pointing to wrong channel, updating")

        // Clear outdated webhooks
        err = discord.WebhookDelete(webhook_id)
        if err != nil {
            return nil, fmt.Errorf("Failed to delete outdated webhook '%s': %v", webhook_id, err)
        }

        // Create new webhook in pin channel
        return createWebhook(discord, guild_id, c.Channel)
    }

    // Return existing webhook if it is valid
    return webhook, nil
}

// createWebhook creates a webhook for a given guild to the given pin channel
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

    err = db.SetWebhook(guild_id, webhook.ID)
    if err != nil {
        return nil, fmt.Errorf("Failed to add webhook to database: %v", err)
    }

    // Update timestamps
    timestamps[webhook.ID] = time.Now().UnixMilli()

    return webhook, nil
}

// splitFiles splits a list of attachments into multiple messages that are under the size limit
func splitFiles(attachments []*discordgo.MessageAttachment, size_limit int) (*[]discordgo.Message, error) {
    file_sets := make([][]*discordgo.File, 0)

    // Reupload attachments if there are any
    for _, a := range msg.Attachments {
        // Download attachment
        file, err := http.DefaultClient.Get(a.URL)
        if err != nil {
            file, err = http.DefaultClient.Get(a.ProxyURL)
            if err != nil {
                return nil, fmt.Errorf("Failed to download attachment '%s': %v", a.URL, err)
            }
        }

        // Reupload attachment
        p.Files = append(p.Files, &discordgo.File{
            Name: a.Filename,
            Reader: file.Body,
        })
    }

    return &p, nil
}

// sizeLimit returns the maximum size (in bytes) of a message that can be sent in a guild
func sizeLimit(discord *discordgo.Session, guild_id string) (int, error) {
    // Get guild object
    guild, err := discord.Guild(guild_id)
    if err != nil {
        return 0, fmt.Errorf("Failed to retrieve guild '%s': %v", guild_id, err)
    }

    var mb int
    switch guild.PremiumTier {
        case 3:
            mb = 100
        case 2:
            mb = 50
        default:
            mb = 25
    }

    return mb * 1024 * 1024, nil
}
