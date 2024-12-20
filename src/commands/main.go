package commands

import (
    "log"
    "fmt"
    "github.com/bwmarrin/discordgo"
)

var index = 0
var signatures = []*discordgo.ApplicationCommand{};

// map[command_name][subcommand_name (if applicable, o.w. "")] = handler
var handlers = map[string]map[string]func(discord *discordgo.Session, option int, i *discordgo.InteractionCreate){}

func RegisterAll(discord *discordgo.Session) error {
    // Populate signature
    registerConfig()
    registerPin()

    // Register signature
    _, err := discord.ApplicationCommandBulkOverwrite(discord.State.User.ID, "", signatures)
    if err != nil {
        return fmt.Errorf("Failed to register main command: %v", err)
    }

    // Register command handler
    discord.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
        // If command is registered
        if _, ok := handlers[i.ApplicationCommandData().Name]; ok {
            // If no options are provided
            if len(i.ApplicationCommandData().Options) == 0 {
                // Execute main command
                handlers[i.ApplicationCommandData().Name][""](s, -1, i)
            } else {
                // For each provided option
                for index, opt := range i.ApplicationCommandData().Options {
                    // If subcommand is registered
                    if cmd, ok := handlers[i.ApplicationCommandData().Name][opt.Name]; ok {
                        // Execute subcommand
                        cmd(s, index, i)
                    }
                }
            }
        }
    })

    log.Printf("Registered %d commands and subcommands", len(handlers))
    return nil;
}

type Command struct {
    metadata *discordgo.ApplicationCommandOption
    handler func(discord *discordgo.Session, option int, i *discordgo.InteractionCreate)
}

func (cmd *Command) register() {
    if cmd.metadata == nil {
        // Add handler for command with no subcommands
        handlers[signatures[index].Name][""] = cmd.handler
    } else {
        // Add handler for command with subcommands
        signatures[index].Options = append(signatures[index].Options, cmd.metadata)
        handlers[signatures[index].Name][cmd.metadata.Name] = cmd.handler
    }
}
