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

var (
    MERGE_TIME int64 = 10*60*1000
    MAX_FILES int = 10
    MAX_LINKS int = 5
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

// cloneMessage recreates the given message into the given webhook with the given base parameters
// Returns the message object that the webhook sent (not including header/footer/attachment messages)
func cloneMessage(discord *discordgo.Session, msg *discordgo.Message, webhook *discordgo.Webhook, base *discordgo.WebhookParams) (*discordgo.Message, error) {
    // Create copy of message as webhook parameters
    params := *base
    params.Content = msg.Content
    params.Components = msg.Components
    params.Embeds = msg.Embeds

    // Append as many attachments to webhook that can fit
    var msgs []*discordgo.WebhookParams

    if len(msg.Attachments) > 0 {
        // Get file upload size limit of guild
        size_limit, err := sizeLimit(discord, webhook.GuildID)
        if err != nil {
            return nil, err
        }

        // Split files based on this size limit
        msgs, err = splitAttachments(base, msg.Attachments, size_limit)
        if err != nil {
            return nil, err
        }

        // Attach first set of files to pin message
        // params.Files = msgs[0].Files
        // msgs = files[1:]
    }

    // Send the webhook copy to the pin channel
    pin_msg, err := discord.WebhookExecute(webhook.ID, webhook.Token, true, &params)
    if err != nil {
        return nil, err
    }

    // Send any attachment messages afterwards
    for _, msg := range msgs {
        _, err := discord.WebhookExecute(webhook.ID, webhook.Token, true, msg)
        if err != nil {
            return nil, err
        }
    }

    return pin_msg, err
}

// splitAttachments splits a list of attachments into multiple messages (as webhook params) under the size limit
func splitAttachments(base *discordgo.WebhookParams, attachments []*discordgo.MessageAttachment, size_limit int) ([]*discordgo.WebhookParams, error) {
    var msgs []*discordgo.WebhookParams
    var file_sets [][]*discordgo.File
    var link_sets [][]string

    current_files := make([]*discordgo.File, 0, MAX_FILES)
    current_links := make([]string, 0, MAX_LINKS)
    current_size := 0
    for _, a := range attachments {
        // Split links if getting too big
        if len(current_links) >= MAX_LINKS {
            link_sets = append(link_sets, current_links)
            current_links = make([]string, 0, MAX_LINKS)
        }

        if a.Size > 0 && a.Size < size_limit {
            // Download attachment
            data, err := http.DefaultClient.Get(a.URL)
            if err != nil {
                data, err = http.DefaultClient.Get(a.ProxyURL)
                if err != nil {
                    // Append link instead if downloading attachment fails
                    current_links = append(current_links, a.URL)
                    continue
                }
            }

            // Split files if getting too big
            if len(current_files) >= MAX_FILES || current_size + a.Size >= size_limit {
                file_sets = append(file_sets, current_files)
                current_files = make([]*discordgo.File, 0, MAX_FILES)
            }

            // Create file with attachment data
            file := &discordgo.File{
                Name: a.Filename,
                ContentType: a.ContentType,
                Reader: data.Body,
            }
            current_files = append(current_files, file)
            current_size += a.Size
        } else {
            // If an attachment is too big to fit in even one message, just append link to it
            current_links = append(current_links, a.URL)
        }
    }

    // Append any remaining files/links
    file_sets = append(file_sets, current_files)
    link_sets = append(link_sets, current_links)

    log.Print("FILES: ", file_sets)
    log.Print("LINKS: ", link_sets)

    return msgs, nil
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
