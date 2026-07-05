package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/nicklaw5/helix/v2"
	"time"
    "log"
	"fmt"
	"strings"
)

func messageDisplayStream(stream helix.Stream) discordgo.MessageSend {
    streamURL := fmt.Sprintf("https://twitch.tv/%s", stream.UserLogin)

    thumbnailURL := strings.NewReplacer(
        "{width}", "1280",
        "{height}", "720",
    ).Replace(stream.ThumbnailURL)
    thumbnailURL += fmt.Sprintf("?t=%d", time.Now().Unix())

    user, err := twitchClient.GetUsers(&helix.UsersParams{
		IDs: []string{stream.UserID},
	})
	if err != nil {
		log.Println("Could not fetch twitch API")
	}
    streamerURL := user.Data.Users[0].ProfileImageURL

    embed := &discordgo.MessageEmbed{
        Thumbnail: &discordgo.MessageEmbedThumbnail{
            URL: streamerURL,
        },
        Author: &discordgo.MessageEmbedAuthor{
            Name: stream.UserName,
            URL:  streamURL,
        },
        Title:       stream.Title,
        URL:         streamURL,
        Description: fmt.Sprintf("🎮 Playing **%s**", stream.GameName),
        Color:       0x9146FF,
        Image: &discordgo.MessageEmbedImage{
            URL: thumbnailURL,
        },
        Fields: []*discordgo.MessageEmbedField{
            {
                Name:   "Viewers",
                Value:  fmt.Sprintf("%d", stream.ViewerCount),
                Inline: true,
            },
            {
                Name:   "Started streaming",
                Value:  fmt.Sprintf("<t:%d:R>", stream.StartedAt.Unix()),
                Inline: true,
            },
        },
        Footer: &discordgo.MessageEmbedFooter{
            Text: "Twitch",
        },
        Timestamp: time.Now().Format(time.RFC3339),
    }

    return discordgo.MessageSend{
        Embeds: []*discordgo.MessageEmbed{embed},
    }
}

func messageError(err error) discordgo.MessageSend { 
	return discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{
			{
				Title: "⚠️ An **error** occured!",
				Color: 0xb5af3a,
				Footer: &discordgo.MessageEmbedFooter{
					Text: time.Now().Format("2006-01-02 15:04:05"),
				},
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "Error",
						Value:  err.Error(),
						Inline: false,
					},
				},
			},
		},
	}
}

func messageCommandError(err error) discordgo.InteractionResponse {
	return discordgo.InteractionResponse{
		Type: 4,
		Data: &discordgo.InteractionResponseData{
			Content: "⚠️ **Error** : " + err.Error(),
		},
	}
}