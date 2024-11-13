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

    // TODO: replace last argument with actual pin message's ID, when thats implemented
    db.AddPin(message.GuildID, message.ID, message.ID)
    if err != nil {
        log.Fatal("Failed to pin message: ", err)
    }

    msgs, err := db.GetMessages(message.GuildID)
    if err != nil {
        log.Fatal("Failed to retrieve messages: ", err)
    }

    for i, msg := range msgs {
        log.Printf("%d -> %s", i, msg)
        discord.ChannelMessageSend(message.ChannelID, fmt.Sprintf("%d -> %s", i, msg))
    }
}
