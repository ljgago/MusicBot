package main

import (
  "sync"
  "github.com/bwmarrin/discordgo"
)

var (
  dg                *discordgo.Session
  voiceInstances    = map[string]*VoiceInstance{}
  purgeTime         int64
  purgeQueue        []PurgeMessage
  mutex             sync.Mutex
  //ignore            = map[string]bool{}
)
