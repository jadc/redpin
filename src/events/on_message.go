package events

import (
	"github.com/bwmarrin/discordgo"
	//"fmt"
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

    c := db.GetConfig(message.GuildID)
    discord.ChannelMessageSend(message.ChannelID, c.Channel)
    if err != nil {
        log.Fatal("Failed to get config: ", err)
    }
    c.Channel = message.Content;
    db.SaveConfig(message.GuildID, c)
}
