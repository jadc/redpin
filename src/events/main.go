package events

import (
    "log"
    "github.com/bwmarrin/discordgo"
)

func RegisterAll(discord *discordgo.Session) {
    // Log bot status
    discord.AddHandler(func(d *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", d.State.User.Username, d.State.User.Discriminator)
	})

    // Register event handlers
    discord.AddHandler(onReaction)
    discord.AddHandler(onReactionRemove)
    discord.AddHandler(onMessageDelete)
    discord.AddHandler(onReady)
    discord.AddHandler(onChannelUpdate)
    discord.AddHandler(onChannelDelete)
}
