package main

import (
  "bufio"
  "encoding/binary"
  "log"
  "io"
  "os/exec"
  "strconv"
  "time"
  "sync"

  "gopkg.in/hraban/opus.v2"
)

// Technically the below settings can be adjusted however that poses
// a lot of other problems that are not handled well at this time.
// These below values seem to provide the best overall performance
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
        ChMessageSend(v.nowPlaying.ChannelID, "[**Music**] End of queue!")
        return
      }
      v.play_wg = &sync.WaitGroup{}
      v.nowPlaying = v.QueueGetSong()
      ChMessageSend(v.nowPlaying.ChannelID, "[**Music**] Playing, **`" + 
        v.nowPlaying.Title + "`  -  `("+ v.nowPlaying.Duration +")`**")
      //dg.UpdateStatus(0, v.nowPlaying.Title)
      pcm := make(chan []int16, 2)
      quit := make(chan bool)
      v.stop = false
      v.skip = false
      v.speaking = true
      v.pause = false
      v.voice.Speaking(true)

      v.play_wg.Add(1)
      go v.SendPCM(pcm, quit, v.play_wg)
      v.play_wg.Add(1)
      go v.SendStream(v.nowPlaying.VideoURL, pcm, quit, v.play_wg)
      v.play_wg.Wait()

      v.QueueRemoveFisrt()
      dg.UpdateStatus(0, o.DiscordStatus)
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
  v.play_wg = &sync.WaitGroup{}
  //dg.UpdateStatus(0, "Radio Streaming")
  pcm := make(chan []int16, 2)
  quit := make(chan bool)
  v.radioFlag = true
  v.stop = false
  v.speaking = true
  v.pause = false
  v.voice.Speaking(true)

  v.play_wg.Add(1)
  go v.SendPCM(pcm, quit, v.play_wg)
  v.play_wg.Add(1)
  go v.SendStream(url, pcm, quit, v.play_wg)
  v.play_wg.Wait()

  dg.UpdateStatus(0, o.DiscordStatus)
  v.radioFlag = false
  v.stop = false
  v.speaking = false
  v.voice.Speaking(false)
}

// SendStream will play the given filename to the already connected
func (v *VoiceInstance) SendStream(url string, pcm chan []int16, quitSig chan bool, wg *sync.WaitGroup) {
  defer wg.Done()
  // Create a shell command "object" to run.
  v.run = exec.Command("ffmpeg", "-i", url, "-f", "s16le", "-ar", strconv.Itoa(frameRate), "-ac", strconv.Itoa(channels), "pipe:1")
  ffmpegout, err := v.run.StdoutPipe()
  if err != nil {
    log.Println("FATA: StdoutPipe Error:", err)
    return
  }
  ffmpegbuf := bufio.NewReaderSize(ffmpegout, 65536)
  // Starts the ffmpeg command
  err = v.run.Start()
  if err != nil {
    log.Println("FATA: RunStart Error:", err)
    return
  }
  defer func() {
    go v.run.Wait()
  }()
  // kill the ffmpeg process
  defer v.run.Process.Kill()
  for {
    // read data from ffmpeg stdout
    select {
      case <-quitSig:
        //quit <- true
        log.Println("INFO: Exit from SendStream.")
        return
      default:  
    }
    if v.stop || v.skip {
      return
    }
    audiobuf := make([]int16, frameSize*channels)
    err = binary.Read(ffmpegbuf, binary.LittleEndian, &audiobuf)
    if err == io.EOF || err == io.ErrUnexpectedEOF {
      log.Println("INFO: Exit from read audio")
      close(pcm)
      return
    } else if err != nil {
      log.Println("INFO: Error reading from ffmpeg stdout :", err)
      close(pcm)
      return
    }
    // Send received PCM to the sendPCM channel
    pcm <- audiobuf
  }
}

// SendPCM will receive on the provied channel encode
// received PCM data into Opus then send that to Discordgo
func (v *VoiceInstance) SendPCM(pcm <-chan []int16, quit chan bool, wg *sync.WaitGroup) {
  var i int
  defer wg.Done()
  opusEncoder, err := opus.NewEncoder(frameRate, channels, opus.AppRestrictedLowdelay)
  if err != nil {
    log.Println("FATA: NewEncoder Error:", err)
    return
  }
  for {
    // read pcm from chan, exit if channel is closed.
    select {
      case <-quit:
        log.Println("INFO: Exit from SendPCM.")
        quit <- true
        return
      case recv, ok := <-pcm:
        if !ok {
          log.Println("INFO: PCM Channel closed.")
          return
        }
        if v.stop || v.skip {
          //quit <- true
          return
        }
        if v.pause {
          //v.lock.Lock()
          for v.pause {
            if v.stop || v.skip {
              break
            }
            time.Sleep(time.Second * 1)
          }
        }
        // try encoding pcm frame with Opus
        opus_data := make([]byte, frameSize*channels*2)
        opus_n, err := opusEncoder.Encode(recv, opus_data)
        if err != nil {
          log.Println("FATA: Encoding Error - ", err)
          //quit <- true
          return
        }
        i = 0
        for v.voice.Ready == false || v.voice.OpusSend == nil {
          log.Printf("FATA: Discordgo not ready for opus packets. %+v : %+v\n", v.voice.Ready, v.voice.OpusSend)
          time.Sleep(1 * time.Second)
          if i > 10 {
            return
          }
          i++
        }
        // send encoded opus data to the send Opus channel
        v.voice.OpusSend <- opus_data[:opus_n] 
    }
  }
}

// Stop stop the audio
func (v *VoiceInstance) Stop() {
  v.stop = true
}

// Pause pause the audio
func (v *VoiceInstance) Pause() {
  v.pause = true
}

// Resume resume the audio
func (v *VoiceInstance) Resume() {
  v.pause = false
}

// KillPlayer forces the player to stop by killing the ffmpeg cmd process
func (v *VoiceInstance) KillPlayer() {
  v.run.Process.Kill()
}

