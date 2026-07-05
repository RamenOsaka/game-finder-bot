package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/nicklaw5/helix/v2"
)

var twitchClient *helix.Client
var configFilePath = "config.json"
var defaultPerms int64 = discordgo.PermissionAdministrator
var guildRuntime = map[string]*GuildRuntime{}

// test server for devs
var testGuildID string = ""