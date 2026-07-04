package main

import (
	"encoding/json"
	"log"
	"os"
	"time"
	"context"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/nicklaw5/helix/v2"
)


func main() {
	// Loading env variables
	discordToken, discordAppID, twitchClientID, twitchClientSecret := loadEnv()
	ctx, cancel := context.WithCancel(context.Background())

	// Setting up twitch application
	var err error

	twitchClient, err = helix.NewClient(&helix.Options{
		ClientID:     twitchClientID,
		ClientSecret: twitchClientSecret,
	})
	if err != nil {
		log.Fatal("Couldn't create client session :", err)
	}
	go maintainAppAccessToken(ctx)

	// Creating new discord session
	dg, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		log.Fatal("Error creating Discord session: ", err)
		return
	}

	//Loading data
	loadConfig()

	// Adding handlers
	dg.AddHandler(ready)
	dg.AddHandler(handlerInteraction)

	// Opening websocket
	err = dg.Open()
	if err != nil {
		log.Println("Error opening Discord session: ", err)
	}

	// Creating commands
	loadCommands(dg, discordAppID, testGuildID)

	log.Println("Game Finder is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	cancel()
	dg.Close()
}

func fetchStreams(s *discordgo.Session, guildID string) {
	resp, err := twitchClient.GetStreams(&helix.StreamsParams{
		GameIDs: []string{guildRuntime[guildID].guildConfig.TwitchGameID},
	})
	if err != nil {
		log.Println("Couldn't fetch streams:", err)
		message := messageError(err)
		s.ChannelMessageSendComplex(guildRuntime[guildID].guildConfig.LogChannel, &message)
	}

	var streamMap = map[string]helix.Stream{}
	for _, stream := range resp.Data.Streams {
		streamMap[stream.UserID] = stream
	}

	for streamerID, messageID := range guildRuntime[guildID].activeStreams {
		if _, exists := streamMap[streamerID]; !exists {
			err := s.ChannelMessageDelete(guildRuntime[guildID].guildConfig.DisplayChannel, messageID)
			if err != nil {
				log.Println("Couldn't delete messages:", err)
				message := messageError(err)
				s.ChannelMessageSendComplex(guildRuntime[guildID].guildConfig.LogChannel, &message)
			}
			config := guildRuntime[guildID]
			delete(config.activeStreams, streamerID)
			guildRuntime[guildID] = config
		}
	}

	for streamerID, stream := range streamMap {
		if _, exists := guildRuntime[guildID].activeStreams[streamerID]; !exists {
			sendMessage := messageDisplayStream(stream)
			message, err := s.ChannelMessageSendComplex(guildRuntime[guildID].guildConfig.DisplayChannel, &sendMessage)
			if err != nil {
				log.Println("Could not send stream message: ", err)
				message := messageError(err)
				s.ChannelMessageSendComplex(guildRuntime[guildID].guildConfig.LogChannel, &message)
			}
			config := guildRuntime[guildID]
			config.activeStreams[streamerID] = message.ID
			guildRuntime[guildID] = config
		}
	}
}

func pollTwitch(ctx context.Context, s *discordgo.Session, guildID string) {
	guildRuntime[guildID].activeStreams = map[string]string{}
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case<-ctx.Done():
			log.Println("Polling stopped:", ctx.Err())
			for _, messageID := range guildRuntime[guildID].activeStreams {
				err := s.ChannelMessageDelete(guildRuntime[guildID].guildConfig.DisplayChannel, messageID)
				if err != nil {
					log.Println("Couldn't delete messages:", err)
					message := messageError(err)
					s.ChannelMessageSendComplex(guildRuntime[guildID].guildConfig.LogChannel, &message)
				}
			}
			return
		case <-ticker.C:
			fetchStreams(s, guildID)
			log.Println("Polling...")
		}
	}
}

func refreshAccessToken() (time.Duration, error) {
	resp, err := twitchClient.RequestAppAccessToken([]string{})
	if err != nil {
		return 0, err
	}
	twitchClient.SetAppAccessToken(resp.Data.AccessToken)
	return time.Duration(resp.Data.ExpiresIn) * time.Second, nil
}

func maintainAppAccessToken(ctx context.Context) {
	for {
		expiresIn, err := refreshAccessToken()
		if err != nil {
			log.Println("Couldn't get/refresh the Twitch App Access Token: ", err)
			return
		}

		refreshIn := expiresIn - time.Hour

		select {
		case<-ctx.Done():
			return
		case<-time.After(refreshIn):
		}
	}
}


func loadCommands(s *discordgo.Session, appID string, guildID string) {
	_, err := s.ApplicationCommandBulkOverwrite(appID, "", []*discordgo.ApplicationCommand{})
	if err != nil {
		log.Println("Could not delete the global commands: ", err)
		message := messageError(err)
		s.ChannelMessageSendComplex(guildRuntime[guildID].guildConfig.LogChannel, &message)
	}
	_, err = s.ApplicationCommandBulkOverwrite(appID, guildID, []*discordgo.ApplicationCommand{})
	if err != nil {
		message := messageError(err)
		s.ChannelMessageSendComplex(guildRuntime[guildID].guildConfig.LogChannel, &message)
		log.Println("Could not delete the guild commands: ", err)
	}

	var applicationCommandList []*discordgo.ApplicationCommand
	for _, cmd := range commands {
		applicationCommandList = append(applicationCommandList, cmd.Definition)
	}
	_, err = s.ApplicationCommandBulkOverwrite(appID, guildID, applicationCommandList)
	if err != nil {
		message := messageError(err)
		s.ChannelMessageSendComplex(guildRuntime[guildID].guildConfig.LogChannel, &message)
		log.Println("Couldn't register commands: ", err)
	}
}

func saveConfig() {
	var permData = map[string]GuildConfig{}
	for key, value := range guildRuntime {
		permData[key] = value.guildConfig
	}
	config, err := json.Marshal(permData)
	if err != nil {
		log.Println("Could not save guild data: ", err)
	}
	os.WriteFile(configFilePath, config, 0644)
}

func loadConfig() {
	var data map[string]GuildConfig
	config, err := os.ReadFile(configFilePath)
	if err != nil {
		log.Println(configFilePath + " Hasn't been created yet : ", err)
		guildRuntime = map[string]*GuildRuntime{}
		return
	} else if len(config) == 0 {
		guildRuntime = map[string]*GuildRuntime{}
		return
	}
	json.Unmarshal(config, &data)

	for key, value := range data {
		guildRuntime[key] = &GuildRuntime{
			guildConfig: value,
		}
	}
}

func loadEnv() (string, string, string, string) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	discordToken, exists := os.LookupEnv("DISCORD_TOKEN")
	if !exists {
		log.Fatal("DISCORD_TOKEN is not set!")
	} else if discordToken == "" {
		log.Fatal("DISCORD_TOKEN is empty!")
	}
	discordAppID, exists := os.LookupEnv("DISCORD_APP_ID")
	if !exists {
		log.Fatal("DISCORD_APP_ID is not set!")
	} else if discordAppID == "" {
		log.Fatal("DISCORD_APP_ID is empty!")
	}

	twitchClientID, exists := os.LookupEnv("TWITCH_CLIENT_ID")
	if !exists {
		log.Fatal("TWITCH_CLIENT_ID is not set!")
	} else if twitchClientID == "" {
		log.Fatal("TWITCH_CLIENT_ID is empty!")
	}
	twitchClientSecret, exists := os.LookupEnv("TWITCH_CLIENT_SECRET")
	if !exists {
		log.Fatal("TWITCH_CLIENT_SECRET is not set!")
	} else if twitchClientSecret == "" {
		log.Fatal("TWITCH_CLIENT_SECRET is empty!")
	}
	return discordToken, discordAppID, twitchClientID, twitchClientSecret
}
