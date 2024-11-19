package commands

import (
    "github.com/bwmarrin/discordgo"
)

var test_command = Command{
    metadata: &discordgo.ApplicationCommand{
        Name: "basic-command",
        Description: "Basic command",
    },
    handler: func(discord *discordgo.Session, i *discordgo.InteractionCreate) {
        discord.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: "Works",
            },
        })
    },
}
