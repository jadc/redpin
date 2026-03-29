package events

import (
	"log"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/jadc/redpin/misc"
)

// Hashmap of channel id to pin count
// Used to prevent attempting to pin when a message is unpinned
var counts = make(map[string]int)
var countsMu sync.Mutex

// hasPinCountIncreased updates the cached pin count for a channel 
// and returns whether the new count is greater than the previous one.
func hasPinCountIncreased(channelID string, count int) bool {
    countsMu.Lock()
    defer countsMu.Unlock()

    prev, ok := counts[channelID]
    counts[channelID] = count

	// If not cached (!ok) then pin anyway, might be a pin decrease
    return !ok || count > prev
}

func onPin(discord *discordgo.Session, event *discordgo.ChannelPinsUpdate) {
    // Get pinned messages in channel
    pins, err := discord.ChannelMessagesPinned(event.ChannelID)
    if err != nil {
        return
    }

    if !hasPinCountIncreased(event.ChannelID, len(pins)) {
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
