package misc

import (
	"database/sql"
	"fmt"
    "log"
    "net/http"
    "strings"

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

// cloneMessage recreates the given message into the given webhook with the given base parameters
// Returns the message object that the webhook sent (not including header/footer/attachment messages)
func cloneMessage(discord *discordgo.Session, msg *discordgo.Message, webhook *discordgo.Webhook, base *discordgo.WebhookParams) (*discordgo.Message, error) {
    var pin_msg *discordgo.Message
    skip := false

    // Create copy of message as webhook parameters
    params := *base
    params.Content = msg.Content
    params.Components = msg.Components

    // Only copy rich embeds, not embeds from links (Discord will add them itself)
    for _, e := range msg.Embeds {
        if e.Type == discordgo.EmbedTypeRich {
            params.Embeds = append(params.Embeds, e)
        }
    }

    // Append as many attachments to webhook that can fit
    var file_sets [][]*discordgo.File
    var link_sets [][]string
    i, j := 0, 0

    if len(msg.Attachments) > 0 {
        // Get file upload size limit of guild
        size_limit, err := sizeLimit(discord, webhook.GuildID)
        if err != nil {
            return nil, err
        }

        // Split files based on this size limit
        file_sets, link_sets = splitAttachments(msg.Attachments, size_limit)

        // Messages must either have content or files to be sent, otherwise Discord errors
        // First, attempt to attach first set of files (not links) to pin message
        if len(file_sets) > 0 && len(file_sets[0]) > 0 {
            params.Files = file_sets[0]
            i += 1
        } else {
            if params.Content == "" {
                if len(link_sets) > 0 && len(link_sets[0]) > 0 {
                    // If message content is empty, try to add first link set
                    params.Content = strings.Join(link_sets[0], "\n")
                    j += 1
                } else {
                    // If there are no link sets, skip the pin message
                    skip = true
                }
            }
        }
    }

    // Send the webhook copy to the pin channel
    if !skip {
        var err error
        pin_msg, err = discord.WebhookExecute(webhook.ID, webhook.Token, true, &params)
        if err != nil {
            return nil, err
        }
    }

    // Send any attachment messages afterwards
    att := *base
    for {
        // Pop from files
        if i < len(file_sets) {
            att.Files = file_sets[i]
            i++
        } else {
            att.Files = nil
        }

        // Pop from links
        if j < len(link_sets) {
            att.Content = strings.Join(link_sets[j], "\n")
            j++
        } else {
            att.Content = ""
        }

        // Send attachment message
        if att.Files != nil || att.Content != "" {
            att_msg, err := discord.WebhookExecute(webhook.ID, webhook.Token, true, &att)
            if err != nil {
                return nil, err
            }

            // If pin message was skipped, set pin message to first attachment message
            if skip {
                pin_msg = att_msg
                skip = false
            }
        } else {
            break
        }
    }

    return pin_msg, nil
}

// splitAttachments splits a list of attachments into list of lists of attachments, each sublist under the size limit
func splitAttachments(attachments []*discordgo.MessageAttachment, size_limit int) ([][]*discordgo.File, [][]string) {
    var file_sets [][]*discordgo.File
    var link_sets [][]string

    files := make([]*discordgo.File, 0, MAX_FILES)
    links := make([]string, 0, MAX_LINKS)
    size := 0

    for _, a := range attachments {
        // Split links if getting too big
        if len(links) >= MAX_LINKS {
            link_sets = append(link_sets, links)
            links = make([]string, 0, MAX_LINKS)
        }

        if a.Size > 0 && a.Size < size_limit {
            // Download attachment
            data, err := http.DefaultClient.Get(a.URL)
            if err != nil {
                data, err = http.DefaultClient.Get(a.ProxyURL)
                if err != nil {
                    // Append link instead if downloading attachment fails
                    links = append(links, a.URL)
                    continue
                }
            }

            // Split files if getting too big
            if len(files) >= MAX_FILES || size + a.Size >= size_limit {
                file_sets = append(file_sets, files)
                files = make([]*discordgo.File, 0, MAX_FILES)
            }

            // Create file with attachment data
            file := &discordgo.File{
                Name: a.Filename,
                ContentType: a.ContentType,
                Reader: data.Body,
            }
            files = append(files, file)
            size += a.Size
        } else {
            // If an attachment is too big to fit in even one message, just append link to it
            links = append(links, a.URL)
        }
    }

    // Append any remaining files/links
    file_sets = append(file_sets, files)
    link_sets = append(link_sets, links)

    return file_sets, link_sets
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

// createReferenceHeader returns a string containing a stylized message reference, used for pins that are replies
func createReferenceHeader(discord *discordgo.Session, webhook *discordgo.Webhook, params *discordgo.WebhookParams, ref *discordgo.MessageReference, depth int) (error) {
    // Get the referenced message
    ref_msg, err := discord.ChannelMessage(ref.ChannelID, ref.MessageID)
    if err != nil {
        return fmt.Errorf("Failed to fetch referenced message: %v", err)
    }

    // Pin the referenced message
    ref_pin_channel_id, ref_pin_msg_id, err := PinMessage(discord, webhook.GuildID, ref_msg, depth)
    if err != nil {
        return fmt.Errorf("Failed to pin referenced message: %v", err)
    }

    // Return formatted link to pinned referenced message
    params.Content = "-# ╰ Reply to " + GetMessageLink(webhook.GuildID, ref_pin_channel_id, ref_pin_msg_id)
    _, err = discord.WebhookExecute(webhook.ID, webhook.Token, true, params)
    if err != nil {
        return fmt.Errorf("Failed to send pin footer: %v", err)
    }

    return nil
}
