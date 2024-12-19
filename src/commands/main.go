package commands

import (
    "log"
    "fmt"
    "github.com/bwmarrin/discordgo"
)

var signature = []*discordgo.ApplicationCommand{
    {
        Name: "redpin",
        Description: "Execute with no arguments to view current config",
        Options: []*discordgo.ApplicationCommandOption{},
    },
};
var handlers = map[string]func(discord *discordgo.Session, i *discordgo.InteractionCreate){}

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
            if options := i.ApplicationCommandData().Options; len(options) == 0 {
                // No arguments
                command_config_main.handler(s, i)
            } else {
                if cmd, ok := handlers[options[0].Name]; ok {
                    cmd(s, i)
                }
            }
        }
    })

    log.Printf("Registered main command and %d subcommands", len(handlers))
    return nil;
}

type Command struct {
    metadata *discordgo.ApplicationCommandOption
    handler func(discord *discordgo.Session, i *discordgo.InteractionCreate)
}

func (cmd *Command) register() {
    signature[0].Options = append(signature[0].Options, cmd.metadata)
    handlers[cmd.metadata.Name] = cmd.handler
}
