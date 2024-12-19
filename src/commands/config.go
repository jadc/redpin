package commands

import (
    "github.com/bwmarrin/discordgo"
)

var command_config_channel = Command{
    metadata: &discordgo.ApplicationCommandOption{
        Name: "channel",
        Description: "Set which channel to send pins to.",
        Type: discordgo.ApplicationCommandOptionChannel,
        ChannelTypes: []discordgo.ChannelType{
            discordgo.ChannelTypeGuildText,
        },
    },
    handler: func(discord *discordgo.Session, i *discordgo.InteractionCreate) {
        discord.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: "channel",
            },
        })
    },
}

var command_config_threshold_min = float64(1);
var command_config_threshold = Command{
    metadata: &discordgo.ApplicationCommandOption{
        Name: "threshold",
        Description: "Set the minimum number of reactions required to pin a message.",
        Type: discordgo.ApplicationCommandOptionInteger,
        MinValue: &command_config_threshold_min,
    },
    handler: func(discord *discordgo.Session, i *discordgo.InteractionCreate) {
        discord.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: "threshold",
            },
        })
    },
}

var command_config_nsfw = Command{
    metadata: &discordgo.ApplicationCommandOption{
        Name: "nsfw",
        Description: "Set whether messages from NSFW channels can be pinned.",
        Type: discordgo.ApplicationCommandOptionBoolean,
    },
    handler: func(discord *discordgo.Session, i *discordgo.InteractionCreate) {
        discord.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: "nsfw",
            },
        })
    },
}

var command_config_selfpin = Command{
    metadata: &discordgo.ApplicationCommandOption{
        Name: "selfpin",
        Description: "Set whether messages can be pinned by their author.",
        Type: discordgo.ApplicationCommandOptionBoolean,
    },
    handler: func(discord *discordgo.Session, i *discordgo.InteractionCreate) {
        discord.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: "self",
            },
        })
    },
}

var command_config_emoji = Command{
    metadata: &discordgo.ApplicationCommandOption{
        Name: "emoji",
        Description: "Customize which emojis can pin messages.",
        Type: discordgo.ApplicationCommandOptionString,
    },
    handler: func(discord *discordgo.Session, i *discordgo.InteractionCreate) {
        discord.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: "moji",
            },
        })
    },
}
