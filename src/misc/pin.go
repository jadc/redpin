package misc

import (
	"database/sql"
	"errors"
	"fmt"
    "log"

	"github.com/bwmarrin/discordgo"
	"github.com/jadc/redpin/database"
)

var (
    ALREADY_PINNED = errors.New("Message is already pinned")

    // Hashset of valid message types
    VALID_MSG_TYPE = map[discordgo.MessageType]struct{}{
        discordgo.MessageTypeDefault: {},
        discordgo.MessageTypeReply: {},
        discordgo.MessageTypeChatInputCommand: {},
        discordgo.MessageTypeThreadStarterMessage: {},
    }
)

// PinMessage pins a message, forwarding it to the pin channel
// Returns the used pin channel ID and pin message's ID if successful
func PinMessage(discord *discordgo.Session, webhook *discordgo.Webhook, msg *discordgo.Message, depth int) (string, string, error) {
    // Skip messages that cannot feasibly be pinned
    if _, ok := VALID_MSG_TYPE[msg.Type]; !ok {
        return "", "", fmt.Errorf("This type of message cannot be pinned")
    }

    // Skip pinning webhooks unless it can be ensured its not a pin
    if msg.WebhookID != "" {
        webhook, err := discord.Webhook(msg.WebhookID)
        if err != nil || webhook == nil || webhook.User == nil || webhook.User.ID == discord.State.User.ID {
            return "", "", fmt.Errorf("Ambiguous webhooks cannot be pinned")
        }
    }

    log.Printf("Pinning message '%s' in guild '%s'", msg.ID, webhook.GuildID)

    db, err := database.Connect()
    if err != nil {
        log.Printf("Failed to connect to database: %v", err)
    }

    // Query database for if message is already pinned
    pin_channel_id, pin_id, err := db.GetPin(webhook.GuildID, msg.ID)

    // Only throw up error if it's an actual error (not just row not found)
    if err != nil && err != sql.ErrNoRows {
        return "", "", fmt.Errorf("Failed to retrieve pin id for message '%s': %v", msg.ID, err)
    }

    // Return existing pin_id if it exists
    if len(pin_channel_id) > 0 && len(pin_id) > 0 {
        return pin_channel_id, pin_id, ALREADY_PINNED
    }

    // Get config
    c := db.GetConfig(webhook.GuildID)

    // Create base webhook params
    params := &discordgo.WebhookParams{
        Username: "Unknown",
        AvatarURL: "",

        // Disable pinging
        AllowedMentions: &discordgo.MessageAllowedMentions{},
    }
    if a := msg.Author; a != nil {
        if member, err := discord.GuildMember(webhook.GuildID, a.ID); err == nil {
            params.Username = GetName(member)
            params.AvatarURL = member.AvatarURL("")
        }
    }

    // If the message being pinned is a reply,
    // pin the referenced message first (as long as reply depth isn't reached)
    if msg.MessageReference != nil && depth + 1 <= c.ReplyDepth {
        err := createReferenceHeader(discord, webhook, params, msg.MessageReference, depth + 1)
        if err != nil {
            log.Printf("Failed to send reference header: %v", err)
        }
    }

    // Send the webhook copy to the pin channel
    pin_msg, err := cloneMessage(discord, msg, webhook, params)
    if err != nil {
        return "", "", fmt.Errorf("Failed to clone pin message: %v", err)
    }

    // Send footer
    params.Content = fmt.Sprintf("-# %s %s", GetMessageLink(webhook.GuildID, msg.ChannelID, msg.ID), msg.Author.Mention())
    _, err = discord.WebhookExecute(webhook.ID, webhook.Token, true, params)
    if err != nil {
        return "", "", fmt.Errorf("Failed to send pin footer: %v", err)
    }

    // Copy reactions from original message if possible
    for _, r := range msg.Reactions {
        _ = discord.MessageReactionAdd(pin_msg.ChannelID, pin_msg.ID, r.Emoji.APIName())
    }

    // Add pin message to database
    err = db.AddPin(webhook.GuildID, c.Channel, msg.ID, pin_msg.ID)
    if err != nil {
        return "", "", fmt.Errorf("Failed to add pin to database: %v", err)
    }

    return pin_msg.ChannelID, pin_msg.ID, nil
}
