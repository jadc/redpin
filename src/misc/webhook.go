package misc

import (
	"database/sql"
	"fmt"
    "log"
    "net/http"

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

        err = db.RemoveWebhooks(guild_id)
        if err != nil {
            return nil, fmt.Errorf("Failed to remove outdated webhooks in database: %v", err)
        }

        // Create new webhook in pin channel
        return createWebhook(discord, guild_id, c.Channel)
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

// cloneParams clones a message into an equivalent a webhook params object
func cloneParams(discord *discordgo.Session, guild_id string, params *discordgo.WebhookParams, msg *discordgo.Message) (*discordgo.WebhookParams, error) {
    // Create copy of given params (shallow is enough)
    p := *params

    // Copy all contents that match
    p.Content = msg.Content
    p.Components = msg.Components
    p.Embeds = msg.Embeds

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
