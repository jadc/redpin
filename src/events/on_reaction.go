package events

import (
	"github.com/bwmarrin/discordgo"
	"log"

	"github.com/jadc/redpin/database"
)

// Map of message id to reaction emoji id to counts
// Used to keep track of reaction counts instead of
// constantly sending requests to the API (although API is used as a fallback)
var reaction_counts = make(map[string]map[string]int)

func onReaction(discord *discordgo.Session, event *discordgo.MessageReactionAdd) {
    reaction := event.MessageReaction

    // Ignore reaction events triggered by self
    if reaction.UserID == discord.State.User.ID {
        return
    }

    // Update reaction count map
    if reaction_counts[event.MessageID] == nil {
        reaction_counts[event.MessageID] = make(map[string]int)
    }
    reaction_counts[event.MessageID][event.Emoji.ID]++
    log.Printf("Added react for message %s, emoji %s, count %d", event.MessageID, event.Emoji.ID, reaction_counts[event.MessageID][event.Emoji.ID])

    // TODO: query database for if message is already pinned

    // Fetch channel
    channel, err := discord.Channel(reaction.ChannelID)
    if err != nil {
        log.Printf("Failed to fetch channel with ID %s: %v", reaction.ChannelID, err)
    }

    // Ignore reactions in NSFW channels
    // TODO: make this configurable
    if channel.NSFW {
        return
    }

    // Fetch message
    message, err := discord.ChannelMessage(reaction.ChannelID, reaction.MessageID)
    if err != nil {
        log.Printf("Failed to fetch message with ID %s: %v", reaction.MessageID, err)
    }

    if !shouldPin(discord, message) {
        return
    }

    db, err := database.Connect()
    if err != nil {
        log.Printf("Failed to connect to database: %v", err)
    }

    // TODO: replace last argument with actual pin event's ID, when thats implemented
    db.AddPin(event.GuildID, event.MessageID, event.MessageID)
    if err != nil {
        log.Fatal("Failed to pin event: ", err)
    }

    log.Printf("Pinned message with ID %s", reaction.MessageID)
}

func onReactionRemove(discord *discordgo.Session, event *discordgo.MessageReactionRemove) {
    // Update reaction count map
    if reaction_counts[event.MessageID] != nil {
        reaction_counts[event.MessageID][event.Emoji.ID]--
        log.Printf("Removed react for message %s, emoji %s, count %d", event.MessageID, event.Emoji.ID, reaction_counts[event.MessageID][event.Emoji.ID])
    }
}

func onMessageDelete(discord *discordgo.Session, event *discordgo.MessageDelete) {
    // Update reaction count map
    if reaction_counts[event.Message.ID] != nil {
        delete(reaction_counts, event.Message.ID)
        log.Printf("Deleted reaction count for message %s", event.Message.ID)
    }
}

// shouldPin checks all reactions of the messsage, and determines if it should be pinned.
func shouldPin(discord *discordgo.Session, message *discordgo.Message) bool {
    for _, r := range message.Reactions {
        // Ignore reactions not in the allowlist
        // TODO: implement this

        count := r.Count

        // Remove reactions from the message author from the count
        // TODO: make this configurable
        /*
        if r.== reaction.UserID {
            return
        }
        */

        // Apply count multiplier for super reacts
        // TODO: implement this, https://discord.com/developers/docs/resources/message#reaction-object-reaction-structure

        // Pin messages with any reactions geq the threshold
        // TODO: make this configurable
        if count >= 2 {
            return true
        }
    }

    return false
}
