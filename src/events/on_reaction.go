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
        log.Printf("Failed to fetch message '%s': %v", reaction.MessageID, err)
        return
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
        return
    }
    c := db.GetConfig(event.GuildID)

    // Ignore reactions in pin channel
    if reaction.ChannelID == c.Channel {
        return
    }

    // Ignore reactions in NSFW channels
    if !c.NSFW && isNSFW(discord, reaction.ChannelID) {
        return
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
        log.Printf("Failed to pin message '%s': %v", message.ID, err)
        return
    }

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
                delete(selfpin[event.MessageID], event.Emoji.ID)
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
            if _, ok := selfpin[message.ID][r.Emoji.ID]; ok {
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

// Hashset of ids of NSFW channels
// Cached to reduce API calls
var is_nsfw = make(map[string]bool)

// isNSFW returns whether a channel is NSFW or not, reading from cache when possible
func isNSFW(discord *discordgo.Session, channel_id string) bool {
    // Attempt to read from cache
    if nsfw, ok := is_nsfw[channel_id]; ok {
        return nsfw
    }

    // Otherwise, query from Discord
    channel, _ := discord.Channel(channel_id)
    is_nsfw[channel_id] = channel.NSFW
    return is_nsfw[channel_id]
}

func onReady(discord *discordgo.Session, event *discordgo.Ready) {
    for _, guild := range event.Guilds {
        channels, _ := discord.GuildChannels(guild.ID)
        for _, channel := range channels {
            is_nsfw[channel.ID] = channel.NSFW
        }
    }
    log.Printf("Cached %d channels' NSFW status", len(is_nsfw))
}

func onChannelUpdate(discord *discordgo.Session, event *discordgo.ChannelUpdate) {
    is_nsfw[event.Channel.ID] = event.NSFW
}

func onChannelDelete(discord *discordgo.Session, event *discordgo.ChannelDelete) {
    delete(is_nsfw, event.Channel.ID)
}
