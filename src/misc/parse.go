package misc

import (
	"fmt"
	"regexp"
	"slices"

	emoji "github.com/Andrew-M-C/go.emoji"
	"github.com/bwmarrin/discordgo"
)

var (
    EMOJI = regexp.MustCompile(`<(a)?:[\w]+:(\d+)>`)
)

// ExtractEmojis returns an identifier for each emoji in the given string
func ExtractEmojis(text string) []string {
    var res []string

    // Extract and append any Discord emojis to result
    if matches := EMOJI.FindAllStringSubmatch(text, -1); matches != nil {
        for _, match := range matches {
            res = append(res, match[2])
        }
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

// GetEmojiID returns an identifier for the given emoji
// If the emoji is a custom Discord emoji, the identifier is the emoji ID
// If the emoji is a unicode emoji, the identifier is the emoji itself
func GetEmojiID(emoji *discordgo.Emoji) string {
    if len(emoji.ID) == 0 {
        return emoji.Name
    }
    return emoji.ID
}

// GetMessageLink returns a URL for the given message
func GetMessageLink(guild_id string, channel_id string, message_id string) string {
    return fmt.Sprintf("https://discord.com/channels/%s/%s/%s", guild_id, channel_id, message_id)
}
