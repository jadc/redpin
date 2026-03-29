package events

import (
	"github.com/bwmarrin/discordgo"
	"log"

	"github.com/jadc/redpin/database"
	"github.com/jadc/redpin/misc"
)

// Map of message id to reaction emoji API name
// Used to prevent self-pinning (if that is disabled)
// Cached to reduce API calls
var selfpin = make(map[string]map[string]struct{})

func onReaction(discord *discordgo.Session, event *discordgo.MessageReactionAdd) {
    reaction := event.MessageReaction

    // Ignore reaction events triggered by self
    if reaction.UserID == discord.State.User.ID {
        return
    }

    message, err := discord.ChannelMessage(reaction.ChannelID, reaction.MessageID)
    if err != nil {
        log.Printf("Failed to fetch message '%s': %v", reaction.MessageID, err)
        return
    }

    // Update selfpin map on reaction add
    if reaction.UserID == message.Author.ID {
        if selfpin[event.MessageID] == nil {
            selfpin[event.MessageID] = make(map[string]struct{})
        }

        selfpin[event.MessageID][event.Emoji.APIName()] = struct{}{}
    }

    db := database.Connect()
    c := db.GetConfig(event.GuildID)

    // Ignore reactions in pin channel
    if reaction.ChannelID == c.Channel {
        return
    }

    // Skip messages that are already pinned
    if _, _, err := db.GetPin(event.GuildID, message.ID); err == nil {
        return
    }

    // Ignore reactions in NSFW channels
    if !c.NSFW {
        channel, err := discord.State.Channel(reaction.ChannelID)
        if err != nil || channel.NSFW {
            return
        }
    }

    if !shouldPin(c, message) {
        return
    }

    // If reaching this far, pin the message
    req, err := misc.CreatePinRequest(discord, event.GuildID, message)
    if err != nil {
        log.Printf("Failed to create pin request for message '%s': %v", message.ID, err)
        return
    }
    misc.Queue.Push(req)

    // Update stats for author of message getting pinned
    err = db.AddStats(event.GuildID, message.Author.ID, reaction.Emoji.MessageFormat())
    if err != nil {
        log.Printf("Failed to update statistics: %v", err)
        return
    }
}

// Update selfpin map on reaction remove
func onReactionRemove(discord *discordgo.Session, event *discordgo.MessageReactionRemove) {
    if selfpin[event.MessageID] != nil {
        reaction := event.MessageReaction
        message, err := discord.ChannelMessage(reaction.ChannelID, reaction.MessageID)
        if err != nil {
            log.Printf("Failed to fetch message '%s': %v", reaction.MessageID, err)
            return
        }

        if reaction.UserID == message.Author.ID {
            if selfpin[event.MessageID] != nil {
                delete(selfpin[event.MessageID], event.Emoji.APIName())
            }
        }
    }
}

// Update selfpin map on message delete
func onMessageDelete(discord *discordgo.Session, event *discordgo.MessageDelete) {
    if selfpin[event.Message.ID] != nil {
        delete(selfpin, event.Message.ID)
    }
}

// shouldPin checks all reactions of the messsage, and determines if the message should be pinned.
func shouldPin(c *database.Config, message *discordgo.Message) bool {
    for _, r := range message.Reactions {
        // If allowlist is non-empty, only then filter emojis
        if len(c.Allowlist) > 0 {
            // Ignore reactions not in the allowlist hashset
            if _, ok := c.Allowlist[r.Emoji.APIName()]; !ok {
                continue
            }
        }

        count := r.Count

        // Remove reactions from the message author from the count
        if !c.Selfpin {
            if _, ok := selfpin[message.ID][r.Emoji.APIName()]; ok {
                count--
            }
        }

        // Pin messages with any reactions geq the threshold
        if count >= c.Threshold {
            return true
        }
    }

    return false
}

