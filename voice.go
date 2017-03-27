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
  run         *exec.Cmd
  stop        bool
)

// PlayAudioFile will play the given filename to the already connected
// Discord voice server/channel.  voice websocket and udp socket
// must already be setup before this will work.
func PlayStream(v *discordgo.VoiceConnection, url string) {
  var pcm chan []int16
  var quit chan bool
  // Create a shell command "object" to run.
  run = exec.Command("ffmpeg", "-i", url, "-f", "s16le", "-ar", strconv.Itoa(frameRate), "-ac", strconv.Itoa(channels), "pipe:1")

  ffmpegout, err := run.StdoutPipe()
  if err != nil {
    log.Println("StdoutPipe Error:", err)
    return
  }

  ffmpegbuf := bufio.NewReaderSize(ffmpegout, 16384)

  // Starts the ffmpeg command
  err = run.Start()
  if err != nil {
    log.Println("RunStart Error:", err)
    return
  }
  // kill the ffmpeg process
  defer run.Process.Kill()

  // Send "speaking" packet over the voice websocket
  v.Speaking(true)
  // Send not "speaking" packet over the websocket when we finish
  defer v.Speaking(false)

  // will actually only spawn one instance, a bit hacky.
  if pcm == nil {
    pcm = make(chan []int16, 2)
  }
  if quit == nil {
    quit = make(chan bool)
  }
  go SendPCM(v, pcm, quit)
  stop = false

  for {
    // read data from ffmpeg stdout
    select {
      case <-quit:
        log.Println("INFO: Exit from PlayAudio.")
        run.Process.Kill()
        return
      default:  
    }

    audiobuf := make([]int16, frameSize*channels)
    err = binary.Read(ffmpegbuf, binary.LittleEndian, &audiobuf)
    if err == io.EOF || err == io.ErrUnexpectedEOF {
      log.Println("FATA: Exit from read audio")
      stop = true
    }
    if err != nil {
      log.Println("FATA: Error reading from ffmpeg stdout :", err)
      stop = true
    }
    if stop == true {
      quit <- true
      <-quit
      log.Println("INFO: Exit from PlayStream")
      stop = false
      return
    }
    // Send received PCM to the sendPCM channel
    pcm <- audiobuf
  }
}

// SendPCM will receive on the provied channel encode
// received PCM data into Opus then send that to Discordgo
func SendPCM(v *discordgo.VoiceConnection, pcm <-chan []int16, quit chan bool) {
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
      log.Println("Encoding Error:", err)
      quit <- true
      return
    }
    count := 0
    for {
      if v.Ready == false || v.OpusSend == nil {
        log.Printf("Discordgo not ready for opus packets. %+v : %+v\n", v.Ready, v.OpusSend)
        time.Sleep(1000 * time.Millisecond)
        if count > 10 {
          quit <- true
          return
        }
        count++
        continue
      } else {
        break
      }
    }
    // send encoded opus data to the send Opus channel
    v.OpusSend <- opus_data[:opus_n]
  }
}

func StopStream() {
  stop = true
}



// KillPlayer forces the player to stop by killing the ffmpeg cmd process
// this method may be removed later in favor of using chans or bools to
// request a stop.
func KillPlayer() {
  run.Process.Kill()
  
}