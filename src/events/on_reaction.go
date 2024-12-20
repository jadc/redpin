package events

import (
	"github.com/bwmarrin/discordgo"
	"log"
	"encoding/json"

	"github.com/jadc/redpin/database"
	"github.com/jadc/redpin/misc"
)

// Map of message id to reaction emoji id
// Used to prevent self-pinning (if that is disabled)
var selfpin = make(map[string]map[string]struct{})

func onReaction(discord *discordgo.Session, event *discordgo.MessageReactionAdd) {
    reaction := event.MessageReaction

    j, _ := json.Marshal(reaction.Emoji)
    log.Print("Reaction detected: ", string(j))

    // Ignore reaction events triggered by self
    if reaction.UserID == discord.State.User.ID {
        return
    }

    message, err := discord.ChannelMessage(reaction.ChannelID, reaction.MessageID)
    if err != nil {
        log.Printf("Failed to fetch message with ID %s: %v", reaction.MessageID, err)
    }

    // Keep track of self pins
    if reaction.UserID == message.Author.ID {
        if selfpin[event.MessageID] == nil {
            selfpin[event.MessageID] = make(map[string]struct{})
        }

        selfpin[event.MessageID][event.Emoji.ID] = struct{}{}
    }

    // TODO: query database for if message is already pinned

    // Fetch channel
    channel, err := discord.Channel(reaction.ChannelID)
    if err != nil {
        log.Printf("Failed to fetch channel with ID %s: %v", reaction.ChannelID, err)
    }

    db, err := database.Connect()
    if err != nil {
        log.Printf("Failed to connect to database: %v", err)
    }
    c := db.GetConfig(event.GuildID)

    // Ignore reactions in NSFW channels
    if channel.NSFW && !c.NSFW {
        return
    }

    if !shouldPin(discord, c, message) {
        return
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
    if selfpin[event.MessageID] != nil {
        reaction := event.MessageReaction
        message, err := discord.ChannelMessage(reaction.ChannelID, reaction.MessageID)
        if err != nil {
            log.Printf("Failed to fetch message with ID %s: %v", reaction.MessageID, err)
        }

        if reaction.UserID == message.Author.ID {
            if selfpin[event.MessageID] != nil {
                delete(selfpin[event.MessageID], event.Emoji.ID)
            }
        }

    }
    log.Printf("Removed react for message %s, emoji %s", event.MessageID, event.Emoji.ID)
}

func onMessageDelete(discord *discordgo.Session, event *discordgo.MessageDelete) {
    // Update reaction count map
    if selfpin[event.Message.ID] != nil {
        delete(selfpin, event.Message.ID)
        log.Printf("Deleted reaction count for message %s", event.Message.ID)
    }
}

// shouldPin checks all reactions of the messsage, and determines if it should be pinned.
func shouldPin(discord *discordgo.Session, c *database.Config, message *discordgo.Message) bool {
    for _, r := range message.Reactions {
        // If allowlist is non-empty, only then filter emojis
        if len(c.Allowlist) > 0 {
            // Ignore reactions not in the allowlist hashset
            if _, ok := c.Allowlist[misc.GetEmojiID(r.Emoji)]; !ok {
                log.Print("Skipping reaction not in allowlist")
                continue
            }
        }

        count := r.Count

        // Remove reactions from the message author from the count
        if !c.Selfpin {
            if _, ok := selfpin[message.ID][r.Emoji.ID]; ok {
                count--
            }
        }

        // Apply count multiplier for super reacts
        // TODO: implement this, https://discord.com/developers/docs/resources/message#reaction-object-reaction-structure

        // Pin messages with any reactions geq the threshold
        log.Printf("Reaction count: %d", count)
        if count >= c.Threshold {
            return true
        }
    }

    return false
}
