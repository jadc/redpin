package commands

import (
	"fmt"
	"log"

    "github.com/bwmarrin/discordgo"
	"github.com/jadc/redpin/misc"
)

func registerPin() error {
    // Add signature
    sig := &discordgo.ApplicationCommand{
        Name: "Pin Message",
        Type: discordgo.MessageApplicationCommand,
    }
    signatures = append(signatures, sig)
    handlers[sig.Name] = make(map[string]func(discord *discordgo.Session, option int, i *discordgo.InteractionCreate))

    // Register commands
    command_pin.register()
    index += 1

    return nil
}

// Command to view current config for the guild
var command_pin = Command{
    metadata: nil,
    handler: func(discord *discordgo.Session, option int, i *discordgo.InteractionCreate) {
        selected_msg := i.ApplicationCommandData().Resolved.Messages[i.ApplicationCommandData().TargetID]

        // Get the current webhook
        webhook, err := misc.GetWebhook(discord, i.GuildID)
        if err == nil {
            _, _, err = misc.PinMessage(discord, webhook, selected_msg)
        }

        resp := "Message pinned."
        if err != nil {
            log.Printf("Failed to pin message: %v", err)
            resp = fmt.Sprintf("Message not pinned.\n```%v```", err)
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

