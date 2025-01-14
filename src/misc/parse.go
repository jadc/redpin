package misc

import (
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
        res = append(res, match.APIName())
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
    return discordgo.EndpointDiscord + "channels/" + guild_id + "/" + channel_id + "/" + message_id
}

// GetName returns the most appropriate available name for a given user
func GetName(member *discordgo.Member) string {
    name := "Unknown"

    if nick := member.Nick; len(nick) > 0 {
        name = nick
    } else if display_name := member.User.GlobalName; len(display_name) > 0 {
        name = display_name
    } else if username := member.User.Username; len(username) > 0 {
        name = username
    }

    return name
}
