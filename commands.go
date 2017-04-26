package main

import (
  "log"
  "time"
  "strconv"
  "strings"
  "github.com/bwmarrin/discordgo"
)

// HelpReporter
func HelpReporter(m *discordgo.MessageCreate) {
  log.Println("INFO:", m.Author.Username, "send 'help'")
  help := "```go\n`Standard Commands List`\n```\n" +
  "**`" + o.DiscordPrefix + "help`** or **`" + o.DiscordPrefix + "h`**  ->  show help commands.\n" +
  "**`" + o.DiscordPrefix + "join`** or **`" + o.DiscordPrefix + "j`**  ->  the bot join in to voice channel.\n" +
  "**`" + o.DiscordPrefix + "leave`** or **`" + o.DiscordPrefix + "l`**  ->  the bot leave the voice channel.\n" +
  "**`" + o.DiscordPrefix + "play`**  ->  play and add a one song in the queue.\n" +
  "**`" + o.DiscordPrefix + "radio`**  ->  play a URL radio.\n" +
  "**`" + o.DiscordPrefix + "stop`**  ->  stop the player and remove the queue.\n" +
  "**`" + o.DiscordPrefix + "skip`**  ->  skip the actual song and play the next song of the queue.\n" +
  "**`" + o.DiscordPrefix + "pause`**  ->  pause the player.\n" +
  "**`" + o.DiscordPrefix + "resume`**  ->  resume the player.\n" +
  "**`" + o.DiscordPrefix + "queue list`**  ->  show the list of song in the queue.\n" +
  "**`" + o.DiscordPrefix + "queue remove `**  ->  remove a song of queue indexed for a ***number***, an ***@User*** or the ***last*** song, i.e. ***"+ o.DiscordPrefix +"queue remove 2***\n" +
  "**`" + o.DiscordPrefix + "queue clean`**  ->  clean all queue.\n" +
  "**`" + o.DiscordPrefix + "youtube`**  ->  search from youtube.\n\n" +
  "```go\n`Owner Commands List`\n```\n" +
  "**`" + o.DiscordPrefix + "ignore`**  ->  ignore commands of a channel.\n" +
  "**`" + o.DiscordPrefix + "unignore`**  ->  unignore commands of a channel.\n"

  ChMessageSend(m.ChannelID, help)
  //ChMessageSendEmbed(m.ChannelID, "Help", help)
}

// JoinReporter
func JoinReporter(v *VoiceInstance, m *discordgo.MessageCreate) {
  log.Println("INFO:", m.Author.Username, "send 'join'")
  voiceChannelID := SearchVoiceChannel(m.Author.ID)
  if voiceChannelID == "" {
    log.Println("ERROR: Voice channel id not found.")
    ChMessageSend(m.ChannelID, "[**Music**] **`"+ m.Author.Username +
      "`** You need to join a voice channel!") 
    return
  }
  if v != nil {
    log.Println("INFO: Voice Instance already created.")
  } else {
    guildID := SearchGuild(m.ChannelID)
    // create new voice instance
    v = new(VoiceInstance)
    mutex.Lock()
    voiceInstances[guildID] = v
    mutex.Unlock()
    v.guildID = guildID
    v.InitVoice()
  }
  var err error
  v.voice, err = dg.ChannelVoiceJoin(v.guildID, voiceChannelID, false, false)
  if err != nil {
    v.Stop()
    log.Println("ERROR: Error to join in a voice channel: ", err)
    return
  }
  v.voice.Speaking(false)
  log.Println("INFO: New Voice Instance created")
  ChMessageSend(m.ChannelID, "[**Music**] I've joined a voice channel!")
}

// LeaveReporter
func LeaveReporter(v *VoiceInstance, m *discordgo.MessageCreate) {
  log.Println("INFO:", m.Author.Username, "send 'leave'")
  if v == nil {
    log.Println("INFO: The bot is not joined a voice channel")
    return
  }
  go func() {
    v.endSig <- true
  }()
  time.Sleep(200 * time.Millisecond)
  v.voice.Disconnect()
  log.Println("INFO: Voice channel destroyed")
  mutex.Lock()
  delete(voiceInstances, v.guildID)
  mutex.Unlock()
  dg.UpdateStatus(0, o.DiscordStatus)
  ChMessageSend(m.ChannelID, "[**Music**] I left the voice channel!")
}

// PlayReporter
func PlayReporter(v *VoiceInstance, m *discordgo.MessageCreate) {
  log.Println("INFO:", m.Author.Username, "send 'play'")
  if v == nil {
    log.Println("INFO: The bot is not joined in voice channel")
    ChMessageSend(m.ChannelID, "[**Music**] I need join in a voice channel!")
    return
  }
  if len(strings.Fields(m.Content)) < 2 {
    ChMessageSend(m.ChannelID, "[**Music**] You need specify an URL.")
    return
  } 
  // send play my_song_youtube
  command := strings.SplitAfter(m.Content, strings.Fields(m.Content)[0])
  query := strings.TrimSpace(command[1])
  song, err := YoutubeFind(query, m)
  if err != nil || song.ID == "" {
    log.Println("ERROR: Youtube search: ", err)
    ChMessageSend(m.ChannelID, "[**Music**] I can't found this song!")
    return
  }
  ChMessageSend(m.ChannelID, "[**Music**] **`"+ song.User +"`** has added , **`"+ 
    song.Title +"`** to the queue. **`("+ song.Duration+ ")` `["+ strconv.Itoa(len(v.queue)) +"]`**")
  go func() {
    v.songSig <- song
  }() 
}

// ReadioReporter
func RadioReporter(v *VoiceInstance, m *discordgo.MessageCreate) {
  log.Println("INFO:", m.Author.Username, "send 'radio'")
  if v == nil {
    log.Println("INFO: The bot is not joined in voice channel")
    ChMessageSend(m.ChannelID, "[**Music**] I need join in a voice channel!")
    return
  }
  if len(strings.Fields(m.Content)) < 2 {
    ChMessageSend(m.ChannelID, "[**Music**] You need to specify a url!")
    return
  }
  radio := strings.Fields(m.Content)[1]
  go func() {
    v.radioSig <- radio
  }()
  ChMessageSend(m.ChannelID, "[**Music**] `"+ m.Author.Username +"` I'm playing a radio now!")
}

// StopReporter
func StopReporter(v *VoiceInstance, m *discordgo.MessageCreate) {
  log.Println("INFO:", m.Author.Username, "send 'stop'")
  if v == nil {
    log.Println("INFO: The bot is not joined in a voice channel")
    ChMessageSend(m.ChannelID, "[**Music**] I need join in a voice channel!")
    return
  }
  v.Stop()
  dg.UpdateStatus(0, o.DiscordStatus)
  log.Println("INFO: The bot stop play audio")
  ChMessageSend(m.ChannelID, "[**Music**] I'm stoped now!")
}

// PauseReporter
func PauseReporter(v *VoiceInstance, m *discordgo.MessageCreate) {
  log.Println("INFO:", m.Author.Username, "send 'pause'")
  if v == nil {
    log.Println("INFO: The bot is not joined in a voice channel")
    return
  }
  if !v.speaking{
    ChMessageSend(m.ChannelID, "[**Music**] I'm not playing nothing!")
    return
  }
  if !v.pause {
    v.Pause()
    ChMessageSend(m.ChannelID, "[**Music**] I'm `PAUSED` now!")
  }
}

// ResumeReporter
func ResumeReporter(v *VoiceInstance, m *discordgo.MessageCreate) {
  log.Println("INFO:", m.Author.Username, "send 'resume'")
  if v == nil {
    log.Println("INFO: The bot is not joined in voice channel")
    ChMessageSend(m.ChannelID, "[**Music**] I need join in a voice channel!")
    return
  }
  if !v.speaking {
    ChMessageSend(m.ChannelID, "[**Music**] I'm not playing nothing!")
    return
  }
  if v.pause {
    v.Resume()
    ChMessageSend(m.ChannelID, "[**Music**] I'm `RESUMED` now!")
  }
}

// QueueReporter
func QueueReporter(v *VoiceInstance, m *discordgo.MessageCreate) {
  log.Println("INFO:", m.Author.Username, "send 'queue'")
  if v == nil {
    log.Println("INFO: The bot is not joined in a voice channel")
    ChMessageSend(m.ChannelID, "[**Music**] I need join in a voice channel!")
    return
  }
  if len(v.queue) == 0 {
    log.Println("INFO: The queue is empty.")
    ChMessageSend(m.ChannelID, "[**Music**] The song queue is empty!")
    return
  }
  if len(strings.Fields(m.Content)) < 2 {
    ChMessageSend(m.ChannelID, "[**Music**] You need specify a `sub-command`!")
    return
  }
  if strings.Contains(m.Content, "queue clean") {
    log.Println("INFO:", m.Author.Username, "send 'queue clean'")
    v.QueueClean()
    ChMessageSend(m.ChannelID, "[**Music**] Queue cleaned")
    return
  }
  if strings.Contains(m.Content, "queue remove") {
    log.Println("INFO:", m.Author.Username, "send 'queue remove'")
    if len(strings.Fields(m.Content)) != 3 {
      ChMessageSend(m.ChannelID, "[**Music**] You need define a `number`, an `@User` or `last` command")
      return
    }
    // is a number?
    if k, err := strconv.Atoi(strings.Fields(m.Content)[2]); err == nil {
      if k < len(v.queue) && k != 0 {
        song := v.queue[k]
        v.QueueRemoveIndex(k)
        ChMessageSend(m.ChannelID, "[**Music**] The songs  **`["+ strconv.Itoa(k) +"]`  -  `"+ song.Title +"`**  was removed of queue!")
        return
      } else {
        ChMessageSend(m.ChannelID, "[**Music**] The songs **`["+ strconv.Itoa(k) +"]`** not exist!")
        return
      }
    }
    // is an user?
    if len(m.Mentions) != 0 {
      v.QueueRemoveUser(m.Mentions[0].Username)
      ChMessageSend(m.ChannelID, "[**Music**] The songs indexed by **`"+ m.Mentions[0].Username +"`** was removed of queue!")
      return
    }
    // the `last` song?
    if strings.Contains(m.Content, "queue remove last") {
      log.Println("INFO:", m.Author.Username, "send 'queue remove last'")
      if len(v.queue) > 1 {
        v.QueueRemoveLast()
        ChMessageSend(m.ChannelID, "[**Music**] The last songs indexed was removed of queue!")
        return
      }
      ChMessageSend(m.ChannelID, "[**Music**] No more songs in the queue!")
      return
    }

  }
  // queue list
  if strings.Contains(m.Content, "queue list") {
    log.Println("INFO:", m.Author.Username, "send 'queue list'")
    message := "[**Music**] My songs are:\n\nNow Playing: **`"+ v.nowPlaying.Title +"`  -  `("+
      v.nowPlaying.Duration +")`  -  `"+ v.nowPlaying.User +"`**\n"
    
    queue := v.queue[1:]
    if len(queue) != 0 {
      var duration TimeDuration
      for i, q := range queue {
        message = message + "\n**`["+ strconv.Itoa(i+1) +"]`  -  `"+ q.Title +"`  -  `("+ q.Duration +")`  -  `"+ q.User +"`**"
        d := strings.Split(q.Duration, ":")
        
        switch (len(d)) {
          case 2:
            // mm:ss
            ss, _ := strconv.Atoi(d[1])
            duration.Second = duration.Second + ss
            mm, _ := strconv.Atoi(d[0])
            duration.Minute = duration.Minute + mm
          case 3:
            // hh:mm:ss
            ss, _ := strconv.Atoi(d[2])
            duration.Second = duration.Second + ss
            mm, _ := strconv.Atoi(d[1])
            duration.Minute = duration.Minute + mm
            hh, _ := strconv.Atoi(d[0])
            duration.Hour = duration.Hour + hh
          case 4:
            // dd:hh:mm:ss
            ss, _ := strconv.Atoi(d[3])
            duration.Second = duration.Second + ss
            mm, _ := strconv.Atoi(d[2])
            duration.Minute = duration.Minute + mm
            hh, _ := strconv.Atoi(d[1])
            duration.Hour = duration.Hour + hh
            dd, _ := strconv.Atoi(d[0])
            duration.Day = duration.Day + dd
        }
      }   
      t := AddTimeDuration(duration)
      message = message + "\n\nThe total duration: **`" +
      strconv.Itoa(t.Day) +"d` `" +
      strconv.Itoa(t.Hour) +"h` `" +
      strconv.Itoa(t.Minute) +"m` `" +
      strconv.Itoa(t.Second) +"s`**"
    }
    ChMessageSend(m.ChannelID, message)
    return
  }
}

// SkipReporter
func SkipReporter(v *VoiceInstance, m *discordgo.MessageCreate) {
  log.Println("INFO:", m.Author.Username, "send 'skip'")
  if v == nil {
    log.Println("INFO: The bot is not joined in voice channel")
    ChMessageSend(m.ChannelID, "[**Music**] I need join in a voice channel!")
    return
  }
  if len(v.queue) == 0 {
    log.Println("INFO: The queue is empty.")
    ChMessageSend(m.ChannelID, "[**Music**] Currently there's no music playing, add some? ;)")
    return
  }
  if v.Skip() {
    ChMessageSend(m.ChannelID, "[**Music**] I'm `PAUSED`, please `resume` first.")
  }
}

// YoutubeReporter
func YoutubeReporter(m *discordgo.MessageCreate) {
  log.Println("INFO:", m.Author.Username, "send 'youtube'")
  command := strings.SplitAfter(m.Content, strings.Fields(m.Content)[0])
  query := strings.TrimSpace(command[1])
  song, err := YoutubeFind(query, m)
  if err != nil || song.ID == "" {
    log.Println("ERROR: Youtube search: ", err)
    ChMessageSend(m.ChannelID, "[**Music**] I can't found this song!")
    return
  }
  ChMessageSendHold(m.ChannelID, "[**Music**] **`"+ song.User +"`**, Youtube URL: https://www.youtube.com/watch?v="+ song.ID)
}

// Not used for now
// StatusReporter
func StatusReporter(m *discordgo.MessageCreate) {
  log.Println("INFO:", m.Author.Username, "send 'status'")
  if len(strings.Fields(m.Content)) < 2 {
    ChMessageSend(m.ChannelID, "[**Music**] You need to specify a status!")
    return
  }
  command := strings.SplitAfter(m.Content, "status")
  status := strings.TrimSpace(command[1])
  dg.UpdateStatus(0, status)
  ChMessageSend(m.ChannelID, "[**Music**] Status: `" + status + "`")
}

// StatusCleanReporter
func StatusCleanReporter(m *discordgo.MessageCreate) {
  log.Println("INFO:", m.Author.Username, "send 'statusclean'")
  dg.UpdateStatus(0, "")
}
