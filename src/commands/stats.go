package commands

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jadc/redpin/database"
	"github.com/jadc/redpin/misc"
)

var (
    LEADERBOARD_NUM_OF_EMOJIS = 3
    STATS_NUM_OF_EMOJIS = 10
)

func registerStats() error {
    // Add signature
    sig := &discordgo.ApplicationCommand{
        Name: "stats",
        Description: "Displays the top 10 members with the most pins",
        Options: []*discordgo.ApplicationCommandOption{},
    }
    signatures = append(signatures, sig)
    handlers[sig.Name] = make(map[string]func(discord *discordgo.Session, option int, i *discordgo.InteractionCreate))

    // Register commands
    command_stats_leaderboard.register()
    command_stats_user.register()
    index += 1

    return nil
}

var command_stats_leaderboard = Command{
    metadata: nil,
    handler: func(discord *discordgo.Session, option int, i *discordgo.InteractionCreate) {

        // Send message acknowledging request
        discord.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{ Embeds: []*discordgo.MessageEmbed{ LoadingEmbed("") } },
        })

        // Connect to database
        db, err := database.Connect()
        if err != nil {
            log.Printf("Failed to connect to database: %v", err)
        }

        lb, err := db.GetLeaderboard(i.GuildID)
        if err != nil {
            log.Printf("Failed to retrieve leaderboard: %v", err)
            return
        }

        embeds := []*discordgo.MessageEmbed{}
        if len(lb) > 0 {
            for n, stats := range lb {
                embed := &discordgo.MessageEmbed{}

                // Set embed header to identity
                if member, err := discord.GuildMember(i.GuildID, stats.UserID); err == nil {
                    embed.Author = &discordgo.MessageEmbedAuthor{
                        Name: fmt.Sprintf("%d. %s (%d total pins)", n+1, misc.GetName(member), stats.Total),
                        IconURL: member.AvatarURL(""),
                    }
                } else {
                    log.Printf("Failed to retrieve member '%s' for embed: %v", stats.UserID, err)
                    continue
                }

                // Create fields for top 3 most used emojis
                var emojis strings.Builder
                num_of_emojis := min(len(stats.Emojis), LEADERBOARD_NUM_OF_EMOJIS)
                emojis.WriteString("-# ")
                for i, e := range stats.Emojis[:num_of_emojis] {
                    emojis.WriteString(fmt.Sprintf("%s x %d", e.EmojiID, e.Count))
                    if i < num_of_emojis - 1 {
                        emojis.WriteString(", ")
                    }
                }
                embed.Description = emojis.String()

                embeds = append(embeds, embed)
            }
        } else {
            embeds = append(embeds, &discordgo.MessageEmbed{
                Title: ":grey_question:  No data",
            })
        }

        // Edit response with state of pin
        discord.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{ Embeds: &embeds })
    },
}


var command_stats_user = Command{
    metadata: &discordgo.ApplicationCommandOption{
        Name: "user",
        Description: "Set to view a more detailed breakdown for a user",
        Type: discordgo.ApplicationCommandOptionUser,
    },

    handler: func(discord *discordgo.Session, option int, i *discordgo.InteractionCreate) {
        embeds := []*discordgo.MessageEmbed{ LoadingEmbed("") }

        // Send message acknowledging request
        discord.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{ Embeds: embeds },
        })

        if user := i.ApplicationCommandData().Options[0].UserValue(discord); user != nil {
            // Connect to database
            db, err := database.Connect()
            if err != nil {
                log.Printf("Failed to connect to database: %v", err)
            }

            // Build embed header
            if member, err := discord.GuildMember(i.GuildID, user.ID); err == nil {
                embeds[0].Author = &discordgo.MessageEmbedAuthor{
                    Name: "Statistics for " + misc.GetName(member),
                    IconURL: member.AvatarURL(""),
                }
            }

            // Build embed contents
            stats, err := db.GetStats(i.GuildID, user.ID)
            if err != nil {
                log.Printf("Failed to retrieve user stats: %v", err)
                return
            }
            embeds[0].Title = fmt.Sprintf("%d total pins", stats.Total)

            var emojis strings.Builder
            num_of_emojis := min(len(stats.Emojis), STATS_NUM_OF_EMOJIS)
            for _, e := range stats.Emojis[:num_of_emojis] {
                emojis.WriteString(fmt.Sprintf("* %s x %d\n", e.EmojiID, e.Count))
            }
            embeds[0].Description = emojis.String()

        } else {
            embeds[0].Title = ":x:  Could not find user"
        }

        // Edit response with state of pin
        discord.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{ Embeds: &embeds })
    },
}

