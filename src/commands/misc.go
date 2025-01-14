package commands

import (
    "github.com/bwmarrin/discordgo"
)

func LoadingEmbed(task string) *discordgo.MessageEmbed {
    t := "Loading..."

    if task != "" {
        t = task
    }

    return &discordgo.MessageEmbed{
        Title: ":hourglass_flowing_sand:  " + t,
    }
}
