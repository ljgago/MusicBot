package main

import (
  "log"
  "io"
  "time"
  "github.com/jonas747/dca"
)

const (
  channels    int = 2         // 1 for mono, 2 for stereo
  frameRate   int = 48000     // audio sampling rate
  frameSize   int = 960       // uint16 size of each audio frame 960/48KHz = 20ms
  bufferSize  int = 1024      // max size of opus data 1K
)

func (v *VoiceInstance) InitVoice() {
  v.songSig = make(chan Song)
  v.radioSig = make(chan string)
  v.endSig = make(chan bool)
  v.speaking = false
  go v.Play(v.songSig, v.radioSig, v.endSig)
}

func (v *VoiceInstance) Play(songSig chan Song, radioSig chan string, endSig chan bool) {
  for {
    select {
      case song := <-songSig:
        if v.radioFlag {
          v.Stop()
          time.Sleep(200 * time.Millisecond)
        }
        go v.PlayQueue(song)
      case radio := <-radioSig:
        v.Stop()
        time.Sleep(200 * time.Millisecond)  
        go v.Radio(radio)
      case end := <-endSig:
        if end == true {
          v.Stop()
          time.Sleep(200 * time.Millisecond)
        }
        return
    }
  }
}

func (v *VoiceInstance) PlayQueue(song Song) {
  // add song to queue
  v.QueueAdd(song)
  if v.speaking {
    // the bot is playing
    return
  }
  go func() {
    v.audioMutex.Lock()
    defer v.audioMutex.Unlock()
    for {
      if len(v.queue) == 0 {
        dg.UpdateStatus(0, o.DiscordStatus)
        ChMessageSend(v.nowPlaying.ChannelID, "[**Music**] End of queue!")
        return
      }
      v.nowPlaying = v.QueueGetSong()
      go ChMessageSend(v.nowPlaying.ChannelID, "[**Music**] Playing, **`" + 
        v.nowPlaying.Title + "`  -  `("+ v.nowPlaying.Duration +")`**")
      // If monoserver
      if o.DiscordPlayStatus {
        dg.UpdateStatus(0, v.nowPlaying.Title)
      }
      v.stop = false
      v.skip = false
      v.speaking = true
      v.pause = false
      v.voice.Speaking(true)

      v.DCA(v.nowPlaying.VideoURL)

      v.QueueRemoveFisrt()
      if v.stop {
        v.QueueRemove()
      }
      v.stop = false
      v.skip = false
      v.speaking = false
      v.voice.Speaking(false)
    }
  }()
}

func (v *VoiceInstance) Radio(url string) {
  v.audioMutex.Lock()
  defer v.audioMutex.Unlock()
  if o.DiscordPlayStatus {
    dg.UpdateStatus(0, "Radio")
  }
  v.radioFlag = true
  v.stop = false
  v.speaking = true
  v.pause = false
  v.voice.Speaking(true)
  
  v.DCA(url)

  dg.UpdateStatus(0, o.DiscordStatus)
  v.radioFlag = false
  v.stop = false
  v.speaking = false
  v.voice.Speaking(false)
}

// DCA
func (v *VoiceInstance) DCA(url string) {
  opts := dca.StdEncodeOptions
  opts.RawOutput = true
  opts.Bitrate = 96
  opts.Application = "lowdelay"

  encodeSession, err := dca.EncodeFile(url, opts)
  if err != nil {
    log.Println("FATA: Failed creating an encoding session: ", err)
  }
  v.encoder = encodeSession
  done := make(chan error)
  stream := dca.NewStream(encodeSession, v.voice, done)
  v.stream = stream
  for {
    select {
    case err := <-done:
      if err != nil && err != io.EOF {
        log.Println("FATA: An error occured", err)
      }
      // Clean up incase something happened and ffmpeg is still running
      encodeSession.Cleanup()
      return
    }
  }
}

// Stop stop the audio
func (v *VoiceInstance) Stop() {
  v.stop = true
  if v.encoder != nil {
    v.encoder.Cleanup()
  }
}

func (v *VoiceInstance) Skip() (bool) {
  if v.speaking {
    if v.pause {
      return true
    } else {
      if v.encoder != nil {
        v.encoder.Cleanup()
      }
    }
  }
  return false
}

// Pause pause the audio
func (v *VoiceInstance) Pause() {
  v.pause = true
  if v.stream != nil {
    v.stream.SetPaused(true)
  }
}

// Resume resume the audio
func (v *VoiceInstance) Resume() {
  v.pause = false
  if v.stream != nil {
    v.stream.SetPaused(false)
  }
}
