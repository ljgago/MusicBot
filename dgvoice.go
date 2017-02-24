package main

import (
  "bufio"
  "encoding/binary"
  "log"
  "io"
  //"os"
  "os/exec"
  "strconv"
  "sync"
  "time"

  "github.com/bwmarrin/discordgo"
  "gopkg.in/hraban/opus.v2"
)

// NOTE: This API is not final and these are likely to change.

// Technically the below settings can be adjusted however that poses
// a lot of other problems that are not handled well at this time.
// These below values seem to provide the best overall performance
const (
  channels  int = 2                   // 1 for mono, 2 for stereo
  frameRate int = 48000               // audio sampling rate
  frameSize int = 960                 // uint16 size of each audio frame 960/48KHz = 20ms
  bufferSize  int = 1024              // max size of opus data 1K
)

var (
  speakers    map[uint32]*opus.Decoder
  opusEncoder *opus.Encoder
  run         *exec.Cmd
  sendpcm     bool
  recvpcm     bool
  //recv        chan *discordgo.Packet
  recv        []int16
  send        chan []int16
  quit        chan bool
  mu          sync.Mutex
)

// SendPCM will receive on the provied channel encode
// received PCM data into Opus then send that to Discordgo
func SendPCM(v *discordgo.VoiceConnection, pcm <-chan []int16, end chan bool) {

  // make sure this only runs one instance at a time.
  //log.Println("Start")
  mu.Lock()
  if sendpcm || pcm == nil {
    mu.Unlock()
    return
  }
  sendpcm = true
  mu.Unlock()

  defer func() { sendpcm = false }()

  var err error

  opusEncoder, err = opus.NewEncoder(frameRate, channels, opus.AppRestrictedLowdelay)

  if err != nil {
    log.Println("NewEncoder Error:", err)
    return
  }

  for {

    // read pcm from chan, exit if channel is closed.
    select {
      case <-end:
        log.Println("Exit from SendPCM.")
        end <- true
        return
      default:  
    }

    recv, ok := <-pcm
    if !ok {
      log.Println("PCM Channel closed.")
      return
    }

    // try encoding pcm frame with Opus
    opus_data := make([]byte, bufferSize)
    opus_n, err := opusEncoder.Encode(recv, opus_data)
    if err != nil {
      log.Println("Encoding Error:", err)
      return
    }
    count := 0
    for {
      if v.Ready == false || v.OpusSend == nil {
        log.Printf("Discordgo not ready for opus packets. %+v : %+v\n", v.Ready, v.OpusSend)
        time.Sleep(1000 * time.Millisecond)
        if count > 10 {
          return
        }
        count++
        continue
      } else {
        break
      }
    }
    // send encoded opus data to the sendOpus channel
    v.OpusSend <- opus_data[:opus_n]
  }
}

// ReceivePCM will receive on the the Discordgo OpusRecv channel and decode
// the opus audio into PCM then send it on the provided channel.

/*func ReceivePCM(v *discordgo.VoiceConnection, c chan *discordgo.Packet) {

  // make sure this only runs one instance at a time.
  mu.Lock()
  if recvpcm || c == nil {
    mu.Unlock()
    return
  }
  recvpcm = true
  mu.Unlock()

  defer func() { sendpcm = false }()
  var err error

  for {

    if v.Ready == false || v.OpusRecv == nil {
      log.Println("Discordgo not ready to receive opus packets. %+v : %+v", v.Ready, v.OpusRecv)
      return
    }

    p, ok := <-v.OpusRecv
    if !ok {
      return
    }

    if speakers == nil {
      speakers = make(map[uint32]*opus.Decoder)
    }

    _, ok = speakers[p.SSRC]
    if !ok {
      speakers[p.SSRC], err = opus.NewDecoder(sampleRate, channels)
      if err != nil {
        log.Println("error creating opus decoder:", err)
        continue
      }
    }

    p.PCM, err = speakers[p.SSRC].Decode(p.Opus, 960, false)
    if err != nil {
      log.Println("Error decoding opus data: ", err)
      continue
    }

    c <- p
  }
}
*/
// PlayAudioFile will play the given filename to the already connected
// Discord voice server/channel.  voice websocket and udp socket
// must already be setup before this will work.
func PlayAudioFile(v *discordgo.VoiceConnection, filename string) {

  // Create a shell command "object" to run.
  run = exec.Command("ffmpeg", "-i", filename, "-f", "s16le", "-ar", strconv.Itoa(frameRate), "-ac", strconv.Itoa(channels), "pipe:1")
  //run = exec.Command("ffmpeg", "-i", filename, "-filter:a", "\"volumedetect\"", "-vn", "-sn",
  //  "-f", "s16le", "-ar", strconv.Itoa(frameRate), "-ac", strconv.Itoa(channels), "pipe:1")
  //cmd = '"' + self.ffmpeg_cmd + '" -i "' + self.input_file + '" -filter:a "volumedetect" -vn -sn -f null ' + nul

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

  // Send "speaking" packet over the voice websocket
  v.Speaking(true)

  // Send not "speaking" packet over the websocket when we finish
  defer v.Speaking(false)

  // will actually only spawn one instance, a bit hacky.
  if send == nil {
    send = make(chan []int16, 2)
  }

  if quit == nil {
    quit = make(chan bool)
  }

  go SendPCM(v, send, quit)

  for {

    // read data from ffmpeg stdout
    audiobuf := make([]int16, frameSize*channels)
    err = binary.Read(ffmpegbuf, binary.LittleEndian, &audiobuf)
    if err == io.EOF || err == io.ErrUnexpectedEOF {
      log.Println("Exit from read audio")
      send <- audiobuf
      quit <- true
      <- quit
      return
    }
    if err != nil {
      log.Println("error reading from ffmpeg stdout :", err)
      quit <- true
      return
    }

    // Send received PCM to the sendPCM channel
    send <- audiobuf
    /*
    select {
      case <- quit:
        log.Println("Exit from PlayAudioFile")
      default:
        log.Println("Sige")
    }
    */
  }
}

// KillPlayer forces the player to stop by killing the ffmpeg cmd process
// this method may be removed later in favor of using chans or bools to
// request a stop.
func KillPlayer() {
  run.Process.Kill()
  
}