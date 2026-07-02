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

	// Setting up twitch application
	var err error

	twitchClient, err = helix.NewClient(&helix.Options{
		ClientID:     twitchClientID,
		ClientSecret: twitchClientSecret,
	})
	if err != nil {
		log.Fatal("Couldn't create client session :", err)
	}

	resp, err := twitchClient.RequestAppAccessToken([]string{})
	if err != nil {
		log.Fatal(err)
	}
	twitchClient.SetAppAccessToken(resp.Data.AccessToken)

	// Creating new discord session
	dg, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		log.Println("Error creating Discord session: ", err)
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
	loadCommands(dg, discordAppID, guildID)

	log.Println("Game Finder is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	dg.Close()
}

func pollTwitch(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case<-ctx.Done():
			log.Println("Polling stopped:", ctx.Err())
			return
		case <-ticker.C:
			log.Println("Polling...")
		}
	}
}


func loadCommands(s *discordgo.Session, appID string, guildID string) {
	_, err := s.ApplicationCommandBulkOverwrite(appID, "", []*discordgo.ApplicationCommand{})
	if err != nil {
		log.Println("Could not delete the global commands: ", err)
	}
	_, err = s.ApplicationCommandBulkOverwrite(appID, guildID, []*discordgo.ApplicationCommand{})
	if err != nil {
		log.Println("Could not delete the server commands commands: ", err)
	}

	var applicationCommandList []*discordgo.ApplicationCommand
	for _, cmd := range commands {
		applicationCommandList = append(applicationCommandList, cmd.Definition)
	}
	_, err = s.ApplicationCommandBulkOverwrite(appID, guildID, applicationCommandList)
	if err != nil {
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
		log.Println("Could not transform guildRuntime into json data: ", err)
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
		log.Println("DISCORD_TOKEN is not set!")
	} else if discordToken == "" {
		log.Println("DISCORD_TOKEN is empty!")
	}
	discordAppID, exists := os.LookupEnv("DISCORD_APP_ID")
	if !exists {
		log.Println("DISCORD_APP_ID is not set!")
	} else if discordAppID == "" {
		log.Println("DISCORD_APP_ID is empty!")
	}

	twitchClientID, exists := os.LookupEnv("TWITCH_CLIENT_ID")
	if !exists {
		log.Println("TWITCH_CLIENT_ID is not set!")
	} else if twitchClientID == "" {
		log.Println("TWITCH_CLIENT_ID is empty!")
	}
	twitchClientSecret, exists := os.LookupEnv("TWITCH_CLIENT_SECRET")
	if !exists {
		log.Println("TWITCH_CLIENT_SECRET is not set!")
	} else if twitchClientSecret == "" {
		log.Println("TWITCH_CLIENT_SECRET is empty!")
	}
	return discordToken, discordAppID, twitchClientID, twitchClientSecret
}
