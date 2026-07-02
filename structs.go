package main

import (
	"github.com/bwmarrin/discordgo"
	"sync"
	"context"
)

type GuildConfig struct {
	DisplayChannel string
	TwitchGameID string
}

type GuildRuntime struct {
	twitchPolling TwitchPolling
	guildConfig GuildConfig
}

type Command struct {
	Definition *discordgo.ApplicationCommand
	Handler func(s *discordgo.Session, i *discordgo.InteractionCreate)
}

type TwitchPolling struct {
	mu sync.Mutex

	ctx context.Context
	cancel context.CancelFunc
}

func (tp *TwitchPolling) tryStart() bool {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	if tp.ctx == nil {
		tp.ctx, tp.cancel = context.WithCancel(context.Background())
		return true
	}
	return false
}

func (tp *TwitchPolling) tryStop() bool {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	if tp.ctx != nil {
		tp.cancel()
		tp.ctx = nil
		tp.cancel = nil
		return true
	}
	return false
}