package main

import (
  "sync"
  "os/exec"
  "github.com/bwmarrin/discordgo"
  "gopkg.in/hraban/opus.v2"

)

type Options struct {
  DiscordToken    string
  DiscordStatus   string
  DiscordPrefix   string
  YoutubeToken    string
}

type TimeDuration struct {
  Day             int
  Hour            int
  Minute          int
  Second          int
}

type Song struct {
  ChannelID       string
  User            string
  ID              string
  //QueueID         string
  Title           string
  Duration        string
  VideoURL        string
}

type PurgeMessage struct {
  ID, ChannelID   string
  TimeSent        int64
}

type VoiceInstance struct {
  voice           *discordgo.VoiceConnection
  //msg             *discordgo.MessageCreate
  play_wg         *sync.WaitGroup
  opusEncoder     *opus.Encoder
  run             *exec.Cmd
  queueMutex      sync.Mutex
  audioMutex      sync.Mutex
  songSig         chan Song
  radioSig        chan string
  endSig          chan bool
  nowPlaying      Song
  queue           []Song
  recv            []int16
  guildID         string
  channelID       string
  speaking        bool
  pause           bool
  stop            bool
  skip            bool
  radioFlag       bool
}
