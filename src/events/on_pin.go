package events

import (
	"github.com/bwmarrin/discordgo"
	"log"

	"github.com/jadc/redpin/misc"
)

// Hashmap of channel id to pin count
// Used to prevent attempting to pin when a message is unpinned
var counts = make(map[string]int)

func onPin(discord *discordgo.Session, event *discordgo.ChannelPinsUpdate) {
    // Get pinned messages in channel
    pins, err := discord.ChannelMessagesPinned(event.ChannelID)
    if err != nil {
        return
    }

    // Abort if the pin event was removal
    if _, ok := counts[event.ChannelID]; !ok {
        counts[event.ChannelID] = len(pins)
    }
    if len(pins) < counts[event.ChannelID] {
        return
    }

    // Pin the message that was just pinned
    real_pin := pins[0]
    if real_pin != nil {
        req, err := misc.CreatePinRequest(discord, event.GuildID, real_pin)
        if err != nil {
            log.Printf("Failed to create pin request for message '%s': %v", real_pin.ID, err)
            return
        }
        misc.Queue.Push(req)
    }
}
