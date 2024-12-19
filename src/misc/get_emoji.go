package misc

import (
	"regexp"
	"slices"

	emoji "github.com/Andrew-M-C/go.emoji"
)

var discord_emoji = regexp.MustCompile(`<(a)?:[\w]+:(\d+)>`)

// ExtractEmojis returns an identifier for each emoji in the given string
func ExtractEmojis(text string) []string {
    var res []string

    // Extract and append any Discord emojis to result
    if matches := discord_emoji.FindAllStringSubmatch(text, -1); matches != nil {
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
