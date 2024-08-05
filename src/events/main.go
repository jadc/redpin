package events

import (
    "github.com/bwmarrin/discordgo"
)

func RegisterAll(discord *discordgo.Session) {
    discord.AddHandler(onMessage)
}
