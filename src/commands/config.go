package commands

import (
    "github.com/bwmarrin/discordgo"
	"github.com/jadc/redpin/database"
	"log"
	"fmt"
	"encoding/json"
)

// Command to view current config for the guild
var command_config_main = Command{
    metadata: nil,
    handler: func(discord *discordgo.Session, i *discordgo.InteractionCreate) {
        // Fetch config for this guild
        db, err := database.Connect()
        if err != nil {
            log.Printf("Failed to connect to database: %v", err)
            return
        }
        c, err := db.GetConfig(i.GuildID)
        if err != nil {
            log.Printf("Failed to get config: %v", err)
            return
        }
        j, err := json.MarshalIndent(c, "", "    ")
        if err != nil {
            log.Printf("Failed to marshal config: %v", err)
            return
        }

        // Respond with success
        discord.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: fmt.Sprintf("**Current Config**\n```json\n%s\n```", string(j)),
                Flags:   discordgo.MessageFlagsEphemeral,
            },
        })
    },
}

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
        // Fetch config for this guild
        db, err := database.Connect()
        if err != nil {
            log.Printf("Failed to connect to database: %v", err)
            return
        }
        c, err := db.GetConfig(i.GuildID)
        if err != nil {
            log.Printf("Failed to get config: %v", err)
            return
        }

        // Write changes to config and save it
        new_value := i.ApplicationCommandData().Options[0].ChannelValue(discord).ID
        c.Channel = new_value
        err = db.SaveConfig(i.GuildID, c)
        if err != nil {
            log.Printf("Failed to save config: %v", err)
            return
        }

        // Respond with success
        discord.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: fmt.Sprintf("Set pin channel to <#%s>", new_value),
                Flags:   discordgo.MessageFlagsEphemeral,
            },
        })
    },
}

var command_config_threshold_min = float64(1)
var command_config_threshold = Command{
    metadata: &discordgo.ApplicationCommandOption{
        Name: "threshold",
        Description: "Set the minimum number of reactions required to pin a message.",
        Type: discordgo.ApplicationCommandOptionInteger,
        MinValue: &command_config_threshold_min,
    },
    handler: func(discord *discordgo.Session, i *discordgo.InteractionCreate) {
        // Fetch config for this guild
        db, err := database.Connect()
        if err != nil {
            log.Printf("Failed to connect to database: %v", err)
            return
        }
        c, err := db.GetConfig(i.GuildID)
        if err != nil {
            log.Printf("Failed to get config: %v", err)
            return
        }

        // Write changes to config and save it
        new_value := int(i.ApplicationCommandData().Options[0].IntValue())
        c.Threshold = new_value
        err = db.SaveConfig(i.GuildID, c)
        if err != nil {
            log.Printf("Failed to save config: %v", err)
            return
        }

        // Respond with success
        discord.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: fmt.Sprintf("Set reaction threshold to %d", new_value),
                Flags:   discordgo.MessageFlagsEphemeral,
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
        // Fetch config for this guild
        db, err := database.Connect()
        if err != nil {
            log.Printf("Failed to connect to database: %v", err)
            return
        }
        c, err := db.GetConfig(i.GuildID)
        if err != nil {
            log.Printf("Failed to get config: %v", err)
            return
        }

        // Write changes to config and save it
        new_value := i.ApplicationCommandData().Options[0].BoolValue()
        c.NSFW = new_value
        err = db.SaveConfig(i.GuildID, c)
        if err != nil {
            log.Printf("Failed to save config: %v", err)
            return
        }

        // Respond with success
        var resp string
        if c.NSFW {
            resp = "Messages in NSFW channels can now be pinned."
        } else {
            resp = "Messages in NSFW channels can no longer be pinned."
        }
        discord.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: resp,
                Flags:   discordgo.MessageFlagsEphemeral,
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
        // Fetch config for this guild
        db, err := database.Connect()
        if err != nil {
            log.Printf("Failed to connect to database: %v", err)
            return
        }
        c, err := db.GetConfig(i.GuildID)
        if err != nil {
            log.Printf("Failed to get config: %v", err)
            return
        }

        // Write changes to config and save it
        new_value := i.ApplicationCommandData().Options[0].BoolValue()
        c.Selfpin = new_value
        err = db.SaveConfig(i.GuildID, c)
        if err != nil {
            log.Printf("Failed to save config: %v", err)
            return
        }

        // Respond with success
        var resp string
        if c.Selfpin {
            resp = "Messages can now be pinned by their author."
        } else {
            resp = "Messages can no longer be pinned by their author."
        }
        discord.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: resp,
                Flags:   discordgo.MessageFlagsEphemeral,
            },
        })
    },
}

// TODO
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
                Content: "TODO",
            },
        })
    },
}
