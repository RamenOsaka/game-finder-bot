package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/nicklaw5/helix/v2"
	"time"
)

func messageDisplayStream(stream helix.Stream) discordgo.MessageSend { 
	return discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{
			{
				Title: "**" + stream.UserName + "** is streaming " + stream.GameName + "!",
				Color: 0xb57a3a,
				Footer: &discordgo.MessageEmbedFooter{
					Text: time.Now().Format("2006-01-02 15:04:05"),
				},
				// Fields: []*discordgo.MessageEmbedField{
				// 	{
				// 		Name:   "Error",
				// 		Value:  "",
				// 		Inline: false,
				// 	},
				//},
			},
		},
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