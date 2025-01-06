package events

import (
	"github.com/bwmarrin/discordgo"
	"log"

	"github.com/jadc/redpin/database"
	"github.com/jadc/redpin/misc"
)

// Map of message id to reaction emoji id
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
        log.Printf("Failed to fetch message with ID %s: %v", reaction.MessageID, err)
    }

    // Update selfpin map on reaction add
    if reaction.UserID == message.Author.ID {
        if selfpin[event.MessageID] == nil {
            selfpin[event.MessageID] = make(map[string]struct{})
        }

        selfpin[event.MessageID][event.Emoji.ID] = struct{}{}
    }

    db, err := database.Connect()
    if err != nil {
        log.Printf("Failed to connect to database: %v", err)
    }
    c := db.GetConfig(event.GuildID)

    // Ignore reactions in NSFW channels
    if !c.NSFW {
        if _, ok := is_nsfw[reaction.ChannelID]; ok {
            log.Print("Skipping reaction in NSFW channel")
            return
        }
    }

    if !shouldPin(c, message) {
        return
    }

    // Get the current webhook
    webhook, err := misc.GetWebhook(discord, event.GuildID)
    if err != nil {
        log.Printf("Failed to retrieve webhook: %v", err)
        return
    }

    // If reaching this far, pin the message
    _, _, err = misc.PinMessage(discord, webhook, message, 0)
    if err != nil {
        log.Printf("Failed to pin message with ID %s: %v", message.ID, err)
    }
}

// Update selfpin map on reaction remove
func onReactionRemove(discord *discordgo.Session, event *discordgo.MessageReactionRemove) {
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

// Update selfpin map on message delete
func onMessageDelete(discord *discordgo.Session, event *discordgo.MessageDelete) {
    if selfpin[event.Message.ID] != nil {
        delete(selfpin, event.Message.ID)
        log.Printf("Deleted reaction count for message %s", event.Message.ID)
    }
}

// shouldPin checks all reactions of the messsage, and determines if the message should be pinned.
func shouldPin(c *database.Config, message *discordgo.Message) bool {
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

// Hashset of ids of NSFW channels
// Cached to reduce API calls
var is_nsfw = make(map[string]struct{})

func onReady(discord *discordgo.Session, event *discordgo.Ready) {
    for _, guild := range event.Guilds {
        channels, _ := discord.GuildChannels(guild.ID)
        for _, channel := range channels {
            if channel.NSFW {
                is_nsfw[channel.ID] = struct{}{}
            }
        }
    }
    log.Printf("Cached %d channels' NSFW status", len(is_nsfw))
}

func onChannelUpdate(discord *discordgo.Session, event *discordgo.ChannelUpdate) {
    if event.NSFW {
        is_nsfw[event.Channel.ID] = struct{}{}
    }
}

func onChannelDelete(discord *discordgo.Session, event *discordgo.ChannelDelete) {
    if event.NSFW {
        delete(is_nsfw, event.Channel.ID)
    }
}
