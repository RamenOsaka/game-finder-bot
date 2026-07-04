package main

import (
	"fmt"
	"log"
	"errors"
	"github.com/bwmarrin/discordgo"
	"github.com/nicklaw5/helix/v2"
)

var commands = map[string]Command{
	"setstreamchannel": {
		Definition: &discordgo.ApplicationCommand{
			Name:                     "setstreamchannel",
			Description:              "Select channel to output streams to.",
			DefaultMemberPermissions: &defaultPerms,
			Contexts: &[]discordgo.InteractionContextType{discordgo.InteractionContextGuild},
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionChannel,
					Name:        "channel",
					Description: "stream channel",
					Required:    true,
				},
			},
		},
		Handler: handleSetStreamChannel,
	},	
	"setlogchannel": {
		Definition: &discordgo.ApplicationCommand{
			Name:                     "setlogchannel",
			Description:              "Select channel to output logs to.",
			DefaultMemberPermissions: &defaultPerms,
			Contexts: &[]discordgo.InteractionContextType{discordgo.InteractionContextGuild},
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
			Contexts: &[]discordgo.InteractionContextType{discordgo.InteractionContextGuild},
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
			Contexts: &[]discordgo.InteractionContextType{discordgo.InteractionContextGuild},
		},
		Handler: handleStartTwitchPolling,
	},
	"stoppolling": {
		Definition: &discordgo.ApplicationCommand{
			Name:                     "stoppolling",
			Description:              "stopts polling twitch for streamed games",
			DefaultMemberPermissions: &defaultPerms,
			Contexts: &[]discordgo.InteractionContextType{discordgo.InteractionContextGuild},
		},
		Handler: handleStopTwitchPolling,
	},
}

func handleStartTwitchPolling(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// get game name
	game, err := twitchClient.GetGames(&helix.GamesParams{
		IDs: []string{guildRuntime[i.GuildID].guildConfig.TwitchGameID},
	})
	if err != nil {
		log.Println(err)
		return fmt.Errorf("Could not fetch twitch: %w", err)
	}
	
	if len(game.Data.Games) == 0 {
		return fmt.Errorf("Game is not found")
	}

	if !guildRuntime[i.GuildID].twitchPolling.tryStart() {
		return errors.New("The game **" + game.Data.Games[0].Name + "** is already being displayed")
	}

	go pollTwitch(guildRuntime[i.GuildID].twitchPolling.ctx, s, i.GuildID)
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: 4,
		Data: &discordgo.InteractionResponseData{
			Content: "Starting to display live streams of **"+ game.Data.Games[0].Name + "**",
		},
	})
	return nil
}

func handleStopTwitchPolling(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if !guildRuntime[i.GuildID].twitchPolling.tryStop() {
		return errors.New("Nothing is being polled at the moment")
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: 4,
		Data: &discordgo.InteractionResponseData{
			Content: "Stopping to display live streams",
		},
	})
	return nil
}

func handleSetGame(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	game, err := twitchClient.GetGames(&helix.GamesParams{
		Names: []string{i.ApplicationCommandData().Options[0].StringValue()},
	})
	if err != nil {
		log.Println(err)
		return fmt.Errorf("Could not fetch twitch: %w", err)
	}

	if len(game.Data.Games) == 0 {
		return errors.New("This game cannot be found")
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
	return nil
}

func handleSetStreamChannel(s *discordgo.Session, i *discordgo.InteractionCreate) error {
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
	return nil
}

func handleSetLogChannel(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	config := guildRuntime[i.GuildID].guildConfig
	config.LogChannel = i.ApplicationCommandData().Options[0].ChannelValue(s).ID
	guildRuntime[i.GuildID].guildConfig = config
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: 4,
		Data: &discordgo.InteractionResponseData{
			Content: "Channel **" + i.ApplicationCommandData().Options[0].ChannelValue(s).Name + "** has been set as the log channel.",
		},
	})
	saveConfig()
	return nil
}