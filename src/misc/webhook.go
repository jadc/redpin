package misc

import (
	"database/sql"
	"fmt"
    "log"

	"github.com/bwmarrin/discordgo"
	"github.com/jadc/redpin/database"
)

var (
    MAX_FILES int = 10
    MAX_LINKS int = 5
)

type WebhookPair struct {
    WebhookA *discordgo.Webhook
    WebhookB *discordgo.Webhook
    LRU bool
}

// Hashmap of guild id -> pair of webhook ids
// Used to prevent many pins in quick succession from being merged into one message
// Technically, if the bot is restarted, the LRU bit is reset and
// messages may get merged, but this is a very unlikely scenario
var webhooks = make(map[string]*WebhookPair)

// GetWebhook returns the appropriate webhook ID for a given guild
func GetWebhook(discord *discordgo.Session, guild_id string) (*discordgo.Webhook, error) {
    db, err := database.Connect()
    if err != nil {
        return nil, fmt.Errorf("Failed to connect to database: %v", err)
    }
    c := db.GetConfig(guild_id)

    // Attempt to read from cache
    if pair, ok := webhooks[guild_id]; ok {
        // Create new webhooks, and delete the old ones, if expired
        if pair.WebhookA.ChannelID != c.Channel || pair.WebhookB.ChannelID != c.Channel {
            discord.WebhookDelete(pair.WebhookA.ID)
            discord.WebhookDelete(pair.WebhookB.ID)

            pair, err = createWebhook(discord, guild_id, c.Channel)
            if err != nil {
                return nil, err
            }
        }

        // Alternate between the two webhooks
        pair.LRU = !pair.LRU
        if pair.LRU {
            return pair.WebhookB, nil
        } else {
            return pair.WebhookA, nil
        }
    }

    // Query database for an existing webhook
    webhook_a_id, webhook_b_id, err := db.GetWebhook(guild_id)

    // Only throw up error if it's an actual error (not just row not found)
    if err != nil && err != sql.ErrNoRows {
        return nil, fmt.Errorf("Failed to retrieve webhook pair for guild '%s': %v", guild_id, err)
    }

    // Create new webhook if one does not exist already
    if err == sql.ErrNoRows || len(webhook_a_id) == 0 || len(webhook_b_id) == 0 {
        _, err := createWebhook(discord, guild_id, c.Channel)
        if err != nil {
            return nil, err
        }
        return GetWebhook(discord, guild_id)
    }

    // Retrieve webhook objects, create a new one if it's invalid
    var webhook_a *discordgo.Webhook
    var webhook_b *discordgo.Webhook
    webhook_a, err = discord.Webhook(webhook_a_id)
    if err == nil {
        webhook_b, err = discord.Webhook(webhook_b_id)
    }
    if err != nil {
        _, err := createWebhook(discord, guild_id, c.Channel)
        if err != nil {
            return nil, err
        }
        return GetWebhook(discord, guild_id)
    }

    // Cache ond return existing webhook pair
    webhooks[guild_id] = &WebhookPair{ WebhookA: webhook_a, WebhookB: webhook_b }
    return GetWebhook(discord, guild_id)
}

// createWebhook creates a webhook pair for a given guild in the given pin channel
func createWebhook(discord *discordgo.Session, guild_id string, channel_id string) (*WebhookPair, error) {
    db, err := database.Connect()
    if err != nil {
        return nil, fmt.Errorf("Failed to connect to database: %v", err)
    }

    // Create webhook A in given channel
    webhookA, err := discord.WebhookCreate(channel_id, "redpin A", "")
    if err != nil {
        return nil, fmt.Errorf("Failed to create webhook A in channel '%s': %v", channel_id, err)
    }

    // Create webhook A in given channel
    webhookB, err := discord.WebhookCreate(channel_id, "redpin B", "")
    if err != nil {
        return nil, fmt.Errorf("Failed to create webhook B in channel '%s': %v", channel_id, err)
    }

    err = db.SetWebhook(guild_id, webhookA.ID, webhookB.ID)
    if err != nil {
        return nil, fmt.Errorf("Failed to add webhook to database: %v", err)
    }

    // Cache new webhook pair
    webhooks[guild_id] = &WebhookPair{ WebhookA: webhookA, WebhookB: webhookB }
    log.Printf("Created new webhook pair for guild '%s'", guild_id)

    return webhooks[guild_id], nil
}
