package misc

import (
    "log"
	"github.com/bwmarrin/discordgo"
	//"github.com/jadc/redpin/database"
)

// PinMessage pins a message, forwarding it to the pin channel
// The force bool, if true, skips most checks
func PinMessage(discord *discordgo.Session, msg *discordgo.Message) (string, error) {
    /*
    db, err := database.Connect()
    if err != nil {
        log.Printf("Failed to connect to database: %v", err)
    }

    // TODO: query database for if message is already pinned

    // TODO: replace last argument with actual pin event's ID, when thats implemented
    db.AddPin(event.GuildID, event.MessageID, event.MessageID)
    if err != nil {
        log.Fatal("Failed to pin event: ", err)
    }
    */

    log.Printf("Pinned message with ID %s", msg.ID)
    return "", nil
}
