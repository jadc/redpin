package events

import (
	"github.com/bwmarrin/discordgo"
	"fmt"
	"log"

	"github.com/jadc/redpin/database"
)

func onMessage(discord *discordgo.Session, message *discordgo.MessageCreate) {
    if message.Author.ID == discord.State.User.ID {
        return
    }

    db, err := database.Connect()
    if err != nil {
        log.Fatal("Failed to connect to database: ", err)
    }

    db.AddMessage(message.Content)

    msgs, err := db.GetMessages()
    if err != nil {
        log.Fatal("Failed to retrieve messages: ", err)
    }

    for i, msg := range msgs {
        discord.ChannelMessageSend(message.ChannelID, fmt.Sprintf("%d -> %s", i, msg))
    }
}
