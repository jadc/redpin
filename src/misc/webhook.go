package misc

import (
	"database/sql"
	"fmt"
    "log"
    "sync"

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
var webhooksMu sync.Mutex

// GetWebhook returns the appropriate webhook for a given guild
func GetWebhook(discord *discordgo.Session, guild_id string) (*discordgo.Webhook, error) {
    webhooksMu.Lock()
    defer webhooksMu.Unlock()

    db := database.Connect()
    c := db.GetConfig(guild_id)

    if pair, ok := webhooks[guild_id]; ok {
		// Recreate cached webhooks if pin channel has changed
        if pair.WebhookA.ChannelID != c.Channel || pair.WebhookB.ChannelID != c.Channel {
            discord.WebhookDelete(pair.WebhookA.ID)
            discord.WebhookDelete(pair.WebhookB.ID)

            var err error
            pair, err = createWebhook(discord, guild_id, c.Channel)
            if err != nil {
                return nil, err
            }
        }

		// Otherwise, use other webhook in cached pair
        return alternateWebhook(pair), nil
    }

	// Fetch webhook pair from database if not cached
    webhook_a_id, webhook_b_id, err := db.GetWebhook(guild_id)
    if err != nil && err != sql.ErrNoRows {
        return nil, fmt.Errorf("Failed to retrieve webhook pair for guild '%s': %v", guild_id, err)
    }

    var pair *WebhookPair

    if err == sql.ErrNoRows || len(webhook_a_id) == 0 || len(webhook_b_id) == 0 {
		// If no webhook pair in database, create new one
        pair, err = createWebhook(discord, guild_id, c.Channel)
        if err != nil {
            return nil, err
        }
    } else {
		// If webhook pair in databse, fetch webhook object and cache it
        var webhook_a *discordgo.Webhook
        var webhook_b *discordgo.Webhook
        webhook_a, err = discord.Webhook(webhook_a_id)
        if err == nil {
            webhook_b, err = discord.Webhook(webhook_b_id)
        }

        if err != nil {
            // If stored webhook IDs are stale/deleted, create new pair
            pair, err = createWebhook(discord, guild_id, c.Channel)
            if err != nil {
                return nil, err
            }
        } else {
            pair = &WebhookPair{ WebhookA: webhook_a, WebhookB: webhook_b }
            webhooks[guild_id] = pair
        }
    }

    return alternateWebhook(pair), nil
}

// alternateWebhook flips the LRU bit and returns the next webhook in the pair.
func alternateWebhook(pair *WebhookPair) *discordgo.Webhook {
    pair.LRU = !pair.LRU
    if pair.LRU {
        return pair.WebhookB
    }
    return pair.WebhookA
}

// createWebhook creates a webhook pair for a given guild in the given pin channel
func createWebhook(discord *discordgo.Session, guild_id string, channel_id string) (*WebhookPair, error) {
    db := database.Connect()

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
