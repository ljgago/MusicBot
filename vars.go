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
  songSignal        chan PkgSong
  radioSignal       chan PkgRadio          
  //ignore            = map[string]bool{}
)
