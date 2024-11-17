package events

import (
    "github.com/bwmarrin/discordgo"
)

func RegisterAll(discord *discordgo.Session) {
    // discord.AddHandler(onMessage)
    discord.AddHandler(onReaction)
    discord.AddHandler(onReactionRemove)
    discord.AddHandler(onMessageDelete)
}
