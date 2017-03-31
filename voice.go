package main

import (
  "bufio"
  "encoding/binary"
  "log"
  "io"
  "os/exec"
  "strconv"
  "time"

  "github.com/bwmarrin/discordgo"
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

var (
  voiceInstances = map[string]*VoiceInstance{}
)


type VoiceInstance struct {
  voice             *discordgo.VoiceConnection
  pusEncoder       *opus.Encoder
  run               *exec.Cmd
  pcm               chan []int16
  end               chan bool
  quit              chan bool
  recv              []int16
  send              chan []int16
  guildID           string
  channelID         string
  skip              bool
  stop              bool
  trackPlaying      bool
}

// PlayAudioFile will play the given filename to the already connected
// Discord voice server/channel.  voice websocket and udp socket
// must already be setup before this will work.
func (vi *VoiceInstance) PlayStream(url string, end chan bool) {
  // Create a shell command "object" to run.
  vi.run = exec.Command("ffmpeg", "-i", url, "-f", "s16le", "-ar", strconv.Itoa(frameRate), "-ac", strconv.Itoa(channels), "pipe:1")

  ffmpegout, err := vi.run.StdoutPipe()
  if err != nil {
    log.Println("StdoutPipe Error:", err)
    return
  }

  //ffmpegbuf := bufio.NewReaderSize(ffmpegout, 16384)
  ffmpegbuf := bufio.NewReaderSize(ffmpegout, 65536)

  // Starts the ffmpeg command
  err = vi.run.Start()
  if err != nil {
    log.Println("RunStart Error:", err)
    return
  }
  // kill the ffmpeg process
  defer vi.run.Process.Kill()

  // Send "speaking" packet over the voice websocket
  vi.voice.Speaking(true)
  // Send not "speaking" packet over the websocket when we finish
  defer vi.voice.Speaking(false)

  // will actually only spawn one instance, a bit hacky.
  if vi.pcm == nil {
    vi.pcm = make(chan []int16, 2)
  }
  //defer close(vi.pcm)
  
  if vi.quit == nil {
    vi.quit = make(chan bool)
  }
  defer close(vi.quit)

  go vi.SendPCM(vi.pcm, vi.quit)
  vi.stop = false

  defer delete(voiceInstances, vi.voice.GuildID)

  for {
    // read data from ffmpeg stdout
    select {
      case <-vi.quit:
        log.Println("INFO: Exit from PlayAudio.")
        vi.run.Process.Kill()
        return
      default:  
    }

    audiobuf := make([]int16, frameSize*channels)
    err = binary.Read(ffmpegbuf, binary.LittleEndian, &audiobuf)
    if err == io.EOF || err == io.ErrUnexpectedEOF {
      log.Println("FATA: Exit from read audio")
      vi.voice.Disconnect()
      return
    } else if err != nil {
      log.Println("FATA: Error reading from ffmpeg stdout :", err)
      vi.voice.Disconnect()
      return
      //vi.run.Process.Kill()
    }
    if vi.stop == true {
      close(vi.pcm)
      vi.quit <- true
      <-vi.quit
      log.Println("INFO: Exit from PlayStream")
      vi.stop = false
      end <- true
      vi.voice.Disconnect()
      return
    }
    // Send received PCM to the sendPCM channel
    vi.pcm <- audiobuf
  }
}

// SendPCM will receive on the provied channel encode
// received PCM data into Opus then send that to Discordgo
func (vi *VoiceInstance) SendPCM(pcm <-chan []int16, quit chan bool) {
  opusEncoder, err := opus.NewEncoder(frameRate, channels, opus.AppRestrictedLowdelay)
  if err != nil {
    log.Println("NewEncoder Error:", err)
    return
  }
  for {
    // read pcm from chan, exit if channel is closed.
    select {
      case <-quit:
        log.Println("INFO: Exit from SendPCM.")
        quit <- true
        return
      default:  
    }
    recv, ok := <-pcm
    if !ok {
      log.Println("INFO: PCM Channel closed.")
      return
    }
    // try encoding pcm frame with Opus
    opus_data := make([]byte, frameSize*channels*2)//bufferSize)
    opus_n, err := opusEncoder.Encode(recv, opus_data)
    if err != nil {
      log.Println("FATA: Encoding Error - ", err)
      quit <- true
      return
    }
    count := 0
    for {
      if vi.voice.Ready == false || vi.voice.OpusSend == nil {
        log.Printf("FATA: Discordgo not ready for opus packets. %+v : %+v\n", vi.voice.Ready, vi.voice.OpusSend)
        time.Sleep(1000 * time.Millisecond)
        if count > 10 {
          vi.quit <- true
          return
        }
        count++
        continue
      } else {
        break
      }
    }
    // send encoded opus data to the send Opus channel
    vi.voice.OpusSend <- opus_data[:opus_n]
  }
}

func (vi *VoiceInstance) StopStream() {
  vi.stop = true
  <-vi.end 
}

// KillPlayer forces the player to stop by killing the ffmpeg cmd process
// this method may be removed later in favor of using chans or bools to
// request a stop.
func (vi *VoiceInstance) KillPlayer() {
  vi.run.Process.Kill()
}