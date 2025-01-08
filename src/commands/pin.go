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
        msg_link := misc.GetMessageLink(i.GuildID, i.ChannelID, selected_msg.ID)

        // Send message acknowledging request
        resp := fmt.Sprintf("-# :hourglass_flowing_sand: Pinning %s...", msg_link)
        discord.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{ Content: resp },
        })
        var pin_channel_id string
        var pin_msg_id string

        // Get the current webhook
        webhook, err := misc.GetWebhook(discord, i.GuildID)
        if err == nil {
            pin_channel_id, pin_msg_id, err = misc.PinMessage(discord, webhook, selected_msg, 0)
        }

        if err != nil {
            log.Printf("Failed to pin message '%s': %v", selected_msg.ID, err)
            resp = fmt.Sprintf("### :x: Failed to pin %s\n```%v```", msg_link, err)
        } else {
            pin_link := misc.GetMessageLink(i.GuildID, pin_channel_id, pin_msg_id)

            pinner := "Someone"
            if i.Member != nil {
                pinner = i.Member.Mention()
            }

            resp = fmt.Sprintf("### :pushpin: %s pinned [a message](%s). See the [pinned message](%s).", pinner, msg_link, pin_link)
        }

        // Edit response with state of pin
        discord.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{ Content: &resp })
    },
}

