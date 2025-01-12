package commands

import (
    "fmt"

    "github.com/bwmarrin/discordgo"
)

func LoadingEmbed(task string) *discordgo.MessageEmbed {
    t := "Loading..."

    if task != "" {
        t = task
    }

    return &discordgo.MessageEmbed{
        Title: fmt.Sprintf(":hourglass_flowing_sand:  %s", t),
    }
}
