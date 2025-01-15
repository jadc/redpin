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
        embeds := []*discordgo.MessageEmbed{ LoadingEmbed(fmt.Sprintf("Pinning %s...", msg_link)) }

        // Send message acknowledging request
        discord.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{ Embeds: embeds },
        })

        // Pin the selected message
        pin_channel_id, pin_msg_id, err := misc.PinMessage(discord, i.GuildID, selected_msg, 0)

        // Send a response based on the state of the pin
        if err != nil {
            log.Printf("Failed to pin message '%s': %v", selected_msg.ID, err)
            embeds[0].Title = ":x:  Failed to pin " + msg_link
            embeds[0].Fields = append(embeds[0].Fields, &discordgo.MessageEmbedField{
                Name: "Reason",
                Value: fmt.Sprintf("```%v```", err),
            })

            // Edit response with state of pin
            discord.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{ Embeds: &embeds })

        } else {
            pin_link := misc.GetMessageLink(i.GuildID, pin_channel_id, pin_msg_id)

            pinner := "Someone"
            if i.Member != nil {
                pinner = i.Member.Mention()
            }

            // Edit response with state of pin
            resp := fmt.Sprintf("### :pushpin: %s pinned [a message](%s). See the [pinned message](%s).", pinner, msg_link, pin_link)
            discord.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{ Content: &resp, Embeds: &[]*discordgo.MessageEmbed{} })
        }
    },
}

