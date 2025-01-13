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

    // Get the current webhook
    webhook, err := misc.GetWebhook(discord, event.GuildID, pins[0].Author.ID)
    if err != nil {
        return
    }

    // Pin the message that was just pinned
    real_pin := pins[0]
    if real_pin != nil {
        _, _, err = misc.PinMessage(discord, webhook, real_pin, 0)
        if err != nil {
            log.Printf("Failed to pin real pin '%s': %v", real_pin.ID, err)
        }
    }
}
