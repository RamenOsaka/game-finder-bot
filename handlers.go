package main

import (

	"github.com/bwmarrin/discordgo"
)

func ready(s *discordgo.Session, event *discordgo.Ready) {
	s.UpdateGameStatus(0, "👁️ looking out for bots")
}

func handlerInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if _, exists := guildRuntime[i.GuildID]; !exists {
		guildRuntime[i.GuildID] = &GuildRuntime{}
	}

	if i.Type == discordgo.InteractionApplicationCommand {
		if cmd, exists := commands[i.ApplicationCommandData().Name]; exists {
			cmd.Handler(s, i)
		}
	}
}