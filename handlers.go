package main

import (

	"github.com/bwmarrin/discordgo"
)

func ready(s *discordgo.Session, event *discordgo.Ready) {
	s.UpdateGameStatus(0, "Looking out for cool game streams")
}

func handlerInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if _, exists := guildRuntime[i.GuildID]; !exists {
		guildRuntime[i.GuildID] = &GuildRuntime{}
	}

	if i.Type == discordgo.InteractionApplicationCommand {
		if cmd, exists := commands[i.ApplicationCommandData().Name]; exists {
			err := cmd.Handler(s, i)
			if err != nil {
				response := messageCommandError(err)
				s.InteractionRespond(i.Interaction, &response)
			}
		}
	}
}