package commands

import (
    "log"
    "fmt"
    "github.com/bwmarrin/discordgo"
)

var signature = []*discordgo.ApplicationCommand{
    {
        Name: "redpin",
        Description: "A Discord app for pinning messages and comparing stats",
        Options: []*discordgo.ApplicationCommandOption{},
    },
};
var handlers = map[string]func(discord *discordgo.Session, i *discordgo.InteractionCreate){}

type Command struct {
    metadata *discordgo.ApplicationCommandOption
    handler func(discord *discordgo.Session, i *discordgo.InteractionCreate)
}

func RegisterAll(discord *discordgo.Session) error {
    // Register all subcommands
    command_config_channel.register()
    command_config_threshold.register()
    command_config_nsfw.register()
    command_config_selfpin.register()
    command_config_emoji.register()

    // Register redpin command signature
    _, err := discord.ApplicationCommandBulkOverwrite(discord.State.User.ID, "", signature)
    if err != nil {
        return fmt.Errorf("Failed to register main command: %v", err)
    }

    // Register command handler
    discord.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
        if i.ApplicationCommandData().Name == signature[0].Name {
            if cmd, ok := handlers[i.ApplicationCommandData().Options[0].Name]; ok {
                cmd(s, i)
            }
        }
    })

    log.Printf("Registered main command and subcommands")
    return nil;
}

func (cmd *Command) register() {
    signature[0].Options = append(signature[0].Options, cmd.metadata)
    handlers[cmd.metadata.Name] = cmd.handler
    log.Printf("Added " + cmd.metadata.Name + " subcommand")
}
