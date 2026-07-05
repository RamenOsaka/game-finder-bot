package main

import (
	"github.com/bwmarrin/discordgo"
	"sync"
	"context"
)

type GuildConfig struct {
	DisplayChannel string
	TwitchGameID string
	LogChannel string
	MinViewers int64
	EnableHistory bool
	EnableMinViewers bool
	StreamerIDBlacklist []string
}

type GuildRuntime struct {
	// map of streamer ID to message ID
	mu sync.RWMutex
	activeStreams map[string]string
	twitchPolling TwitchPolling
	guildConfig GuildConfig
}

type Command struct {
	Definition *discordgo.ApplicationCommand
	Handler func(s *discordgo.Session, i *discordgo.InteractionCreate) error
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

func (gr *GuildRuntime) Config() GuildConfig {
	gr.mu.RLock()
	defer gr.mu.RUnlock()
	return gr.guildConfig
}

func (gr *GuildRuntime) SetConfig(update func(*GuildConfig)) {
	gr.mu.Lock()
	defer gr.mu.Unlock()
	update(&gr.guildConfig)
}