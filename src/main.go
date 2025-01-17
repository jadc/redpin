package main

import (
    "os"
    "log"

    "github.com/bwmarrin/discordgo"
    "github.com/jadc/redpin/events"
    "github.com/jadc/redpin/commands"
    "github.com/jadc/redpin/misc"
)

func main(){
    // Retrieve token
    token := os.Getenv("DISCORD_TOKEN")
    if token == "" {
        log.Fatal("Environmental variable 'DISCORD_TOKEN' is missing.")
    }

    // Create Discord app instance
    discord, err := discordgo.New("Bot " + token)
    if err != nil {
        log.Fatal("Failed to create Discord session: ", err)
    }

    // Register event handlers
    events.RegisterAll(discord)

    // Open session
    err = discord.Open()
    if err != nil {
        log.Fatal("Failed to start Discord session: ", err)
    }
    defer discord.Close()

    // Register slash commands
    err = commands.RegisterAll(discord)
    if err != nil {
        log.Print("Failed to register custom commands: ", err)
    }

    // Constantly check and consume new pin requests
    misc.Queue = misc.NewQueue()
    for {
        log.Print("Listening for pin requests...")
        _, _, err := misc.Queue.Execute(discord)
        if err != nil {
            log.Print(err)
        }
    }
}
