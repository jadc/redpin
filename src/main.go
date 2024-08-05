package main

import (
	"os"
	"log"
	"os/signal"
	"github.com/bwmarrin/discordgo"

	"strings"
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
    discord.AddHandler(onMessage)

    // Open session
    err = discord.Open()
    if err != nil {
        log.Fatal("Failed to start Discord session: ", err)
    }
    defer discord.Close()

    // Keep thread running until CTRL + C
    block := make(chan os.Signal, 1)
    signal.Notify(block, os.Interrupt)
    log.Println("redpin is online. Press CTRL + C to exit.")
    <-block
    log.Println("Shutting down gracefully...")
}

func onMessage(discord *discordgo.Session, message *discordgo.MessageCreate) {
    if message.Author.ID == discord.State.User.ID {
        return
    }

    if strings.Contains(message.Content, "skibidi") {
        discord.ChannelMessageSend(message.ChannelID, "toilet")
    }
}
