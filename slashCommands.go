package main

import (
	"fmt"
	"log"
	"errors"
	"slices"
	"strconv"
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
	"addblacklist": {
		Definition: &discordgo.ApplicationCommand{
			Name:                     "addblacklist",
			Description:              "add a user to the streams blacklist",
			DefaultMemberPermissions: &defaultPerms,
			Contexts: &[]discordgo.InteractionContextType{discordgo.InteractionContextGuild},
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "user",
					Description: "the user's login name, can be found in the url at USER : https://www.twitch.tv/<USER>",
					Required:    true,
				},
			},
		},
		Handler: handleAddStreamerBlacklist,
	},
	"removeblacklist": {
		Definition: &discordgo.ApplicationCommand{
			Name:                     "removeblacklist",
			Description:              "removes a user from the streams blacklist",
			DefaultMemberPermissions: &defaultPerms,
			Contexts: &[]discordgo.InteractionContextType{discordgo.InteractionContextGuild},
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "user",
					Description: "the user's login name, can be found in the url at USER : https://www.twitch.tv/<USER>",
					Required:    true,
				},
			},
		},
		Handler: handleAddStreamerBlacklist,
	},
	"enablehistory": {
		Definition: &discordgo.ApplicationCommand{
			Name:                     "enablehistory",
			Description:              "enable or disable the twitch streams history in the stream channel",
			DefaultMemberPermissions: &defaultPerms,
			Contexts: &[]discordgo.InteractionContextType{discordgo.InteractionContextGuild},
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "boolean",
					Description: "whether or not you want the history to be enabled",
					Required:    true,
				},
			},
		},
		Handler: handleEnableHistory,
	},
	"enableviewerfloor": {
		Definition: &discordgo.ApplicationCommand{
			Name:                     "enableviewerfloor",
			Description:              "enable or disable whether or not streams are filtered by amount of followers",
			DefaultMemberPermissions: &defaultPerms,
			Contexts: &[]discordgo.InteractionContextType{discordgo.InteractionContextGuild},
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "boolean",
					Description: "whether or not viewer floor is be enabled",
					Required:    true,
				},
			},
		},
		Handler: handleEnableViewerFloor,
	},
	"setminviewers": {
		Definition: &discordgo.ApplicationCommand{
			Name:                     "setminviewers",
			Description:              "set the minimum amount of viewers for the streamer's stream to be displayed",
			DefaultMemberPermissions: &defaultPerms,
			Contexts: &[]discordgo.InteractionContextType{discordgo.InteractionContextGuild},
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "amount",
					Description: "amount of viewers (default to 0)",
					Required:    true,
				},
			},
		},
		Handler: handleSetViewerFloor,
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
		IDs: []string{guildRuntime[i.GuildID].Config().TwitchGameID},
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

func handleEnableHistory(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	config := guildRuntime[i.GuildID].guildConfig
	config.EnableHistory = i.ApplicationCommandData().Options[0].BoolValue()
	guildRuntime[i.GuildID].guildConfig = config
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: 4,
		Data: &discordgo.InteractionResponseData{
			Content: "Stream history is now set to **" + strconv.FormatBool(i.ApplicationCommandData().Options[0].BoolValue()) + "**.",
		},
	})
	saveConfig()
	return nil
}

func handleAddStreamerBlacklist(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	resp, err := twitchClient.GetUsers(&helix.UsersParams{
		Logins: []string{i.ApplicationCommandData().Options[0].StringValue()},
	})
	if err != nil {
		return errors.New("Could not fetch twitch API")
	}
	if len(resp.Data.Users) == 0 {
		return errors.New("This user cannot be found")
	}

	config := guildRuntime[i.GuildID].guildConfig
	config.StreamerIDBlacklist = append(config.StreamerIDBlacklist, resp.Data.Users[0].ID)
	guildRuntime[i.GuildID].guildConfig = config
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: 4,
		Data: &discordgo.InteractionResponseData{
			Content: "Streamer **" + resp.Data.Users[0].DisplayName + "** has been added to the blacklist.",
		},
	})
	saveConfig()
	return nil
}

func handleRemoveStreamerBlacklist(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	resp, err := twitchClient.GetUsers(&helix.UsersParams{
		Logins: []string{i.ApplicationCommandData().Options[0].StringValue()},
	})
	if err != nil {
		return errors.New("Could not fetch twitch API")
	}
	if len(resp.Data.Users) == 0 {
		return errors.New("This user cannot be found")
	}

	config := guildRuntime[i.GuildID].guildConfig
	streamerIndex := slices.Index(config.StreamerIDBlacklist, resp.Data.Users[0].ID)
	if streamerIndex == -1 {
		return errors.New("This user is not part of the blacklist")
	}

	config.StreamerIDBlacklist = slices.Delete(config.StreamerIDBlacklist, streamerIndex, streamerIndex + 1)
	guildRuntime[i.GuildID].guildConfig = config
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: 4,
		Data: &discordgo.InteractionResponseData{
			Content: "Streamer **" + resp.Data.Users[0].DisplayName + "** has been removed from the blacklist.",
		},
	})
	saveConfig()
	return nil
}

func handleEnableViewerFloor(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	config := guildRuntime[i.GuildID].guildConfig
	config.EnableMinViewers = i.ApplicationCommandData().Options[0].BoolValue()
	guildRuntime[i.GuildID].guildConfig = config
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: 4,
		Data: &discordgo.InteractionResponseData{
			Content: "Filtering by followers floor is now set to **" + strconv.FormatBool(i.ApplicationCommandData().Options[0].BoolValue()) + "**.",
		},
	})
	saveConfig()
	return nil
}

func handleSetViewerFloor(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	config := guildRuntime[i.GuildID].guildConfig
	config.MinViewers = i.ApplicationCommandData().Options[0].IntValue()
	guildRuntime[i.GuildID].guildConfig = config
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: 4,
		Data: &discordgo.InteractionResponseData{
			Content: "Followers floor is now set to **" + strconv.FormatInt(i.ApplicationCommandData().Options[0].IntValue(), 10) + "**.",
		},
	})
	saveConfig()
	return nil
}