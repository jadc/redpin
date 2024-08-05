package events

import (
	"github.com/bwmarrin/discordgo"
	"strings"
)

func onMessage(discord *discordgo.Session, message *discordgo.MessageCreate) {
    if message.Author.ID == discord.State.User.ID {
        return
    }

    if strings.Contains(strings.ToLower(message.Content), "skibidi") {
        discord.ChannelMessageSend(message.ChannelID, "toilet")
    }
}
