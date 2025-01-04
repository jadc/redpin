package misc

import (
	"database/sql"
	"fmt"
    "log"
    "net/http"
    "time"
    "encoding/json"

	"github.com/bwmarrin/discordgo"
	"github.com/jadc/redpin/database"
)

var (
    MERGE_TIME int64 = 10*60*1000
)

// Hashmap of webhook id -> last used timestamp
// Used to prevent many pins in quick succession from being merged into one message
// Technically, if the bot is restarted, timestamps are reset and
// messages may get merged, but this is a very unlikely scenario
var timestamps = make(map[string]int64)

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

    // Create new webhook if one does not exist already
    if err == sql.ErrNoRows || webhook_id == "" {
        log.Printf("Creating new webhook for guild '%s'", guild_id)
        return createWebhook(discord, guild_id, c.Channel)
    }

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

// splitFiles splits a list of attachments into lists of files that are each under the size limit
func splitFiles(attachments []*discordgo.MessageAttachment, size_limit int) ([][]*discordgo.File, []string, error) {
    // Preallocate space based on attachments
    total_size := 0
    total_links := 0
    for _, a := range attachments {
        if a.Size >= 0 || a.Size < size_limit {
            total_size += a.Size
        } else {
            total_links += 1
        }
    }
    files := make([][]*discordgo.File, total_size / size_limit + 1)
    links := make([]string, 0, total_links)

    log.Print("Message requires ", len(files), " files and ", len(links), " links")

    // Reupload attachments if there are any
    size_so_far := 0
    for _, a := range attachments {
        // If an attachment is too big to fit in even one message, just copy URL to it
        if a.Size < 0 || a.Size > size_limit {
            links = append(links, a.URL)
            continue
        }

        // Reupload attachment if it has a valid size
        data, err := http.DefaultClient.Get(a.URL)
        if err != nil {
            data, err = http.DefaultClient.Get(a.ProxyURL)
            if err != nil {
                log.Print("Failed to download attachment '%s': %v", a.URL, err)
                links = append(links, a.URL)
            }
        }

        // Create file with attachment data
        file := &discordgo.File{
            Name: a.Filename,
            ContentType: a.ContentType,
            Reader: data.Body,
        }

        // Add file to appropriate set
        size_so_far += a.Size
        i := size_so_far / size_limit
        files[i] = append(files[i], file)
    }

    j, _ := json.MarshalIndent(files, "", "  ")
    log.Print(string(j))
    j, _ = json.MarshalIndent(links, "", "  ")
    log.Print(string(j))

    return files, links, nil
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
