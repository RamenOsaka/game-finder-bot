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
	"startpolling": {
		Definition: &discordgo.ApplicationCommand{
			Name:                     "startpolling",
			Description:              "starts polling twitch for streamed games",
			DefaultMemberPermissions: &defaultPerms,
		},
		Handler: handleStartTwitchPolling,
	},
	"stoppolling": {
		Definition: &discordgo.ApplicationCommand{
			Name:                     "stoppolling",
			Description:              "stopts polling twitch for streamed games",
			DefaultMemberPermissions: &defaultPerms,
		},
		Handler: handleStopTwitchPolling,
	},
}

func handleStartTwitchPolling(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// get game name
	game, err := twitchClient.GetGames(&helix.GamesParams{
		IDs: []string{guildRuntime[i.GuildID].guildConfig.TwitchGameID},
	})
	if err != nil {
		log.Println(err)
	}
	
	if len(game.Data.Games) == 0 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: 4,
			Data: &discordgo.InteractionResponseData{
				Content: "The game hasn't been set up yet.",
			},
		})
		return
	}

	if !guildRuntime[i.GuildID].twitchPolling.tryStart() {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: 4,
			Data: &discordgo.InteractionResponseData{
				Content: "The game **" + game.Data.Games[0].Name + "** is already being displayed",
			},
		})
		return
	}

	go pollTwitch(guildRuntime[i.GuildID].twitchPolling.ctx)
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: 4,
		Data: &discordgo.InteractionResponseData{
			Content: "Starting to display live streams of **"+ game.Data.Games[0].Name + "**",
		},
	})
}

func handleStopTwitchPolling(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// get game name
	game, err := twitchClient.GetGames(&helix.GamesParams{
		IDs: []string{guildRuntime[i.GuildID].guildConfig.TwitchGameID},
	})
	if err != nil {
		log.Println(err)
	}
	
	if len(game.Data.Games) == 0 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: 4,
			Data: &discordgo.InteractionResponseData{
				Content: "The game hasn't been set up yet.",
			},
		})
	}

	if !guildRuntime[i.GuildID].twitchPolling.tryStop() {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: 4,
			Data: &discordgo.InteractionResponseData{
				Content: "Nothing is being polled at the moment.",
			},
		})
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: 4,
		Data: &discordgo.InteractionResponseData{
			Content: "Stopping to display live streams of **"+ game.Data.Games[0].Name + "**",
		},
	})
}

func handleSetGame(s *discordgo.Session, i *discordgo.InteractionCreate) {
	game, err := twitchClient.GetGames(&helix.GamesParams{
		Names: []string{i.ApplicationCommandData().Options[0].StringValue()},
	})
	if err != nil {
		log.Println(err)
	}

	if len(game.Data.Games) == 0 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: 4,
		Data: &discordgo.InteractionResponseData{
			Content: "Game doesn't exist",
		},
	})
	} else {
		config := guildRuntime[i.GuildID].guildConfig
		config.TwitchGameID = game.Data.Games[0].ID
		guildRuntime[i.GuildID].guildConfig = config

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: 4,
			Data: &discordgo.InteractionResponseData{
				Content: "Game **" + game.Data.Games[0].Name + "** has been set.",
			},
		})
	}
	saveConfig()
}

func handleSetLogChannel(s *discordgo.Session, i *discordgo.InteractionCreate) {
	config := guildRuntime[i.GuildID].guildConfig
	config.DisplayChannel = i.ApplicationCommandData().Options[0].ChannelValue(s).ID
	guildRuntime[i.GuildID].guildConfig = config
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: 4,
		Data: &discordgo.InteractionResponseData{
			Content: "Channel **" + i.ApplicationCommandData().Options[0].ChannelValue(s).Name + "** has been set as the display channel.",
		},
	})
	saveConfig()
}
