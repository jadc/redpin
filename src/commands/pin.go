package commands

import (
    "github.com/bwmarrin/discordgo"
	"fmt"
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
        _, err := misc.PinMessage(discord, i.GuildID, selected_msg)

        // Only respond if failure
        if err != nil {
            discord.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
                Type: discordgo.InteractionResponseChannelMessageWithSource,
                Data: &discordgo.InteractionResponseData{
                    Content: fmt.Sprintf("Failed to pin message: %v", err),
                    Flags:   discordgo.MessageFlagsEphemeral,
                },
            })
        }
    },
}

