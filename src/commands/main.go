package commands

import (
    "log"
    "fmt"
    "github.com/bwmarrin/discordgo"
)

type Command struct {
    metadata *discordgo.ApplicationCommand
    handler func(discord *discordgo.Session, i *discordgo.InteractionCreate)
}

func RegisterAll(discord *discordgo.Session) error {
    if err := test_command.register(discord); err != nil { return err }
    return nil;
}

func (cmd *Command) register(discord *discordgo.Session) error {
    // Register command signature
    _, err := discord.ApplicationCommandCreate(discord.State.User.ID, "", cmd.metadata)
    if err != nil {
        return fmt.Errorf("Failed to create '%v' command: %v", cmd.metadata.Name, err)
    }

    // Register command handler
    discord.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
        if i.ApplicationCommandData().Name == cmd.metadata.Name {
            cmd.handler(s, i)
        }
    })

    log.Printf("Registered command '%v'", cmd.metadata.Name)
    return nil;
}

func DeregisterAll(discord *discordgo.Session) error {
    cmds, err := discord.ApplicationCommands(discord.State.User.ID, "")
    if err != nil {
        return fmt.Errorf("Failed to fetch registered commands: %v", err)
    }

    for _, cmd := range cmds {
        err := discord.ApplicationCommandDelete(discord.State.User.ID, "", cmd.ID)
        if err != nil {
            log.Printf("Failed to delete '%v' command: %v", cmd.Name, err)
        }
        log.Printf("Deregistered command '%v'", cmd.Name)
    }

    return nil;
}
