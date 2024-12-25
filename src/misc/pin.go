package misc

import (
	"database/sql"
	"fmt"
    "log"

	"github.com/bwmarrin/discordgo"
	"github.com/jadc/redpin/database"
)

// PinMessage pins a message, forwarding it to the pin channel
// Returns the pin message's ID if successful
func PinMessage(discord *discordgo.Session, guild_id string, msg *discordgo.Message) (string, string, error) {
    log.Printf("Pinning message '%s' in guild '%s'", msg.ID, guild_id)

    db, err := database.Connect()
    if err != nil {
        log.Printf("Failed to connect to database: %v", err)
    }

    // Query database for if message is already pinned
    pin_channel_id, pin_id, err := db.GetPin(guild_id, msg.ID)

    // Only throw up error if it's an actual error (not just row not found)
    if err != nil && err != sql.ErrNoRows {
        return "", "", fmt.Errorf("Failed to retrieve pin id for message with ID %s: %v", msg.ID, err)
    }

    // Return existing pin_id if it exists
    if len(pin_channel_id) > 0 && len(pin_id) > 0 {
        log.Printf("Message with ID %s is already pinned", msg.ID)
        return pin_channel_id, pin_id, nil
    }

    // Get config
    c := db.GetConfig(guild_id)

    // Get the current webhook
    webhook, err := GetWebhook(discord, guild_id)
    if err != nil {
        return "", "", fmt.Errorf("Failed to retrieve webhook: %v", err)
    }

    // Get message author's name
    name, err := getUserName(discord, guild_id, msg.Author)
    if err != nil {
        return "", "", fmt.Errorf("Failed to get username: %v", err)
    }

    // If the message being pinned is a reply, pin the referenced message first
    if msg.MessageReference != nil {
        // Get the referenced message
        ref_msg, err := discord.ChannelMessage(msg.MessageReference.ChannelID, msg.MessageReference.MessageID)
        if err != nil {
            log.Printf("Failed to get referenced message of message '%s': %v", msg.ID, err)
        }

        // Pin the referenced message
        ref_pin_channel_id, ref_pin_msg_id, err := PinMessage(discord, guild_id, ref_msg)
        if err != nil {
            log.Printf("Failed to pin referenced message of message '%s': %v", msg.ID, err)
        }

        // Send link to pinned referenced message
        _, err = discord.WebhookExecute(webhook.ID, webhook.Token, false, &discordgo.WebhookParams{
            Content: fmt.Sprintf("-# Reply to https://discord.com/channels/%s/%s/%s", guild_id, ref_pin_channel_id, ref_pin_msg_id),
            Username: name,
            AvatarURL: msg.Author.AvatarURL(""),
        })
        if err != nil {
            log.Printf("Failed to send link to pinned referenced message of message '%s': %v", msg.ID, err)
        }
    }

    // Create a webhook copy of the message
    // TODO: make deep copy of original message
    params := &discordgo.WebhookParams{
        // Copy identity
        Username: name,
        AvatarURL: msg.Author.AvatarURL(""),

        // Prevent pins from pinging
        AllowedMentions: &discordgo.MessageAllowedMentions{},

        // Copy message content
        Content: msg.Content,
        Embeds: msg.Embeds,
    }

    // Send the webhook copy to the pin channel
    pin_msg, err := discord.WebhookExecute(webhook.ID, webhook.Token, true, params)
    if err != nil {
        return "", "", fmt.Errorf("Failed to execute webhook: %v", err)
    }

    // Add pin message to database
    err = db.AddPin(guild_id, c.Channel, msg.ID, pin_msg.ID)
    if err != nil {
        return "", "", fmt.Errorf("Failed to add pin to database: %v", err)
    }

    // Send link to original message
    _, err = discord.WebhookExecute(webhook.ID, webhook.Token, false, &discordgo.WebhookParams{
        Content: fmt.Sprintf("-# https://discord.com/channels/%s/%s/%s", guild_id, msg.ChannelID, msg.ID),
        Username: name,
        AvatarURL: msg.Author.AvatarURL(""),
    })
    if err != nil {
        return "", "", fmt.Errorf("Failed to execute webhook: %v", err)
    }

    return pin_msg.ChannelID, pin_msg.ID, nil
}
