package main

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/nicklaw5/helix/v2"
)

var commands = map[string]Command{
	"setlogchannel": {
		Definition: &discordgo.ApplicationCommand{
			Name:                     "setlogchannel",
			Description:              "Select channel to output logs to.",
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
	"setgame": {
		Definition: &discordgo.ApplicationCommand{
			Name:                     "setgame",
			Description:              "set the game to watch for",
			DefaultMemberPermissions: &defaultPerms,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "game",
					Description: "game name",
					Required:    true,
				},
			},
		},
		Handler: handleSetGame,
	},
}

func handleSetGame(s *discordgo.Session, i *discordgo.InteractionCreate) {
	games, err := twitchClient.GetGames(&helix.GamesParams{
		Names: []string{i.ApplicationCommandData().Options[0].StringValue()},
	})
	if err != nil {
		log.Fatal(err)
	}

	if len(games.Data.Games) == 0 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: 4,
		Data: &discordgo.InteractionResponseData{
			Content: "Game doesn't exist",
		},
	})
	} else {
		config := serverConfigs[i.GuildID]
		config.TwitchGameID = games.Data.Games[0].ID
		serverConfigs[i.GuildID] = config

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: 4,
			Data: &discordgo.InteractionResponseData{
				Content: "Game **" + games.Data.Games[0].Name + "** has been set.",
			},
		})
	}
	saveConfig()
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
