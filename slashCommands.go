package main

import (
	"github.com/bwmarrin/discordgo"
)

var commands = map[string]Command{
	"setlogchannel": {
		Definition: &discordgo.ApplicationCommand{
			Name: "setlogchannel",
			Description: "Select channel to output logs to.",
			DefaultMemberPermissions: &defaultPerms,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionChannel,
					Name:        "channel",
					Description: "log channel",
					Required:    true,
				},
			},
		},
		Handler: handleSetLogChannel,
	},
}

func handleSetLogChannel(s *discordgo.Session, i *discordgo.InteractionCreate) {
	config := serverConfigs[i.GuildID]
	config.DisplayChannel = i.ApplicationCommandData().Options[0].ChannelValue(s).ID
	serverConfigs[i.GuildID] = config
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: 4,
			Data: &discordgo.InteractionResponseData{
				Content: "Channel **" + i.ApplicationCommandData().Options[0].ChannelValue(s).Name + "** has been set as the display channel.",
			},
		})
	saveConfig()
}