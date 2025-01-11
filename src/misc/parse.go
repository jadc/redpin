package misc

import (
	"fmt"
	"slices"

	emoji "github.com/Andrew-M-C/go.emoji"
	"github.com/bwmarrin/discordgo"
)

// ExtractEmojis returns an identifier for each emoji in the given string
func ExtractEmojis(text string) []string {
    var res []string

    // Extract and append any Discord emojis to result
    temp := &discordgo.Message{ Content: text }
    for _, match := range temp.GetCustomEmojis() {
        res = append(res, match.MessageFormat())
    }

    // Extract and append any unicode emojis to result
    for i := emoji.IterateChars(text); i.Next(); {
        if i.CurrentIsEmoji() {
            res = append(res, i.Current())
        }
    }

    // Return extracted emojis, without duplicates
    slices.Sort(res)
    return slices.Compact(res)
}

// GetMessageLink returns a URL for the given message
func GetMessageLink(guild_id string, channel_id string, message_id string) string {
    return fmt.Sprintf("%schannels/%s/%s/%s", discordgo.EndpointDiscord, guild_id, channel_id, message_id)
}
