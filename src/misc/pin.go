package misc

import (
	"database/sql"
	"fmt"
    "log"

	"github.com/bwmarrin/discordgo"
	"github.com/jadc/redpin/database"
)

// PinMessage pins a message, forwarding it to the pin channel
// Returns the used pin channel ID and pin message's ID if successful
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

    // Create base webhook params
    params := &discordgo.WebhookParams{
        Username: "Unknown",
        AvatarURL: "",

        // Disable pinging
        AllowedMentions: &discordgo.MessageAllowedMentions{},
    }
    if a := msg.Author; a != nil {
        // Get message author's name
        name := a.Username
        member, err := discord.GuildMember(guild_id, a.ID)
        if err == nil {
            if len(member.Nick) > 0 {
                name = member.Nick
            } else if len(member.User.GlobalName) > 0 {
                name = member.User.GlobalName
            }

            params.Username = name
        }

        // Copy avatar
        params.AvatarURL = msg.Author.AvatarURL("")
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
        params.Content = fmt.Sprintf("-# ╰ Reply to https://discord.com/channels/%s/%s/%s", guild_id, ref_pin_channel_id, ref_pin_msg_id)
        _, err = discord.WebhookExecute(webhook.ID, webhook.Token, false, params)
        if err != nil {
            log.Printf("Failed to send link to pinned referenced message of message '%s': %v", msg.ID, err)
        }
    }

    // Send the webhook copy to the pin channel
    pin_msg, err := cloneMessage(discord, msg, webhook, params)
    if err != nil {
        return "", "", fmt.Errorf("Failed to execute webhook: %v", err)
    }

    // Send footer
    params.Content = fmt.Sprintf(
        "-# _ _\n-# https://discord.com/channels/%s/%s/%s %s",
        guild_id, msg.ChannelID, msg.ID, msg.Author.Mention(),
    )
    _, err = discord.WebhookExecute(webhook.ID, webhook.Token, false, params)
    if err != nil {
        return "", "", fmt.Errorf("Failed to execute webhook: %v", err)
    }

    // Copy reactions from original message if possible
    for _, r := range msg.Reactions {
        _ = discord.MessageReactionAdd(pin_msg.ChannelID, pin_msg.ID, r.Emoji.APIName())
    }

    // Add pin message to database
    err = db.AddPin(guild_id, c.Channel, msg.ID, pin_msg.ID)
    if err != nil {
        return "", "", fmt.Errorf("Failed to add pin to database: %v", err)
    }

    return pin_msg.ChannelID, pin_msg.ID, nil
}
