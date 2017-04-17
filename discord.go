package main

import (
  "log"
  "time"
  "strings"
  "github.com/bwmarrin/discordgo"
)


// DiscordConnect make a new connection to Discord
func DiscordConnect() (err error) {
  dg, err = discordgo.New("Bot " + o.DiscordToken)
  if err != nil {
    log.Println("FATA: error creating Discord session,", err)
    return
  }
  log.Println("INFO: Bot is Opening")
  dg.AddHandler(MessageCreateHandler)
  dg.AddHandler(GuildCreateHandler)
  dg.AddHandler(GuildDeleteHandler)
  // Open Websocket
  err = dg.Open()
  if err != nil {
    log.Println("FATA: Error Open():", err)
    return
  }
  _, err = dg.User("@me")
  if err != nil {
    // Login unsuccessful
    log.Println("FATA:", err)
    return
  } // Login successful
  //d.Session.UpdateStatus(0, "")
  log.Println("INFO: Bot user test")
  //log.Println("INFO: Playing", o.Status)
  //log.Println("INFO:", o.Url)
  log.Println("INFO: Bot is now running. Press CTRL-C to exit.")
  // Purge time of 30 sec
  purgeTime = 60
  purgeRoutine()
  dg.UpdateStatus(0, o.DiscordStatus)
  
  return nil
}

// Start connect to discord and join to channel
func Start() (err error) {
  if err = DiscordConnect(); err != nil {
    log.Println("FATA:", err)
    return err
  }
  return nil
}

// SearchVoiceChannel search the voice channel id into from guild.
func SearchVoiceChannel(user string) (voiceChannelID string) { 
  for _, g := range dg.State.Guilds {
    for _, v := range g.VoiceStates {
      if v.UserID == user {
        return v.ChannelID
      }
    }   
  }
  return ""
}

// SearchGuild search the guild ID
func SearchGuild(textChannelID string) (guildID string) {
  channel, _ := dg.Channel(textChannelID)
  guildID = channel.GuildID
  return 
}

// AddTimeDuration calculate the total time duration
func AddTimeDuration(t TimeDuration) (total TimeDuration) {
  total.Second =  t.Second % 60
  t.Minute = t.Minute + t.Second / 60
  total.Minute = t.Minute % 60
  t.Hour = t.Hour + t.Minute / 60
  total.Hour = t.Hour % 24
  total.Day = t.Day + t.Hour / 24
  return
}

// ChMessageSendEmbed
func ChMessageSendEmbed(textChannelID, title, description string) {
  embed := discordgo.MessageEmbed{}
  embed.Title = title
  embed.Description = description
  embed.Color = 0xb20000
  for i := 0; i < 10; i++ {
    msg, err := dg.ChannelMessageSendEmbed(textChannelID, &embed)
    if err != nil {
      time.Sleep(1 * time.Second)
      continue
    }
    msgToPurgeQueue(msg)
    break
  }
}

// ChMessageSendHold send a message
func ChMessageSendHold(textChannelID, message string) {
  for i := 0; i < 10; i++ {
    _, err := dg.ChannelMessageSend(textChannelID, message)
    if err != nil {
      time.Sleep(1 * time.Second)
      continue
    }
    break
  }
}

// ChMessageSend send a message and auto-remove it in a time
func ChMessageSend(textChannelID, message string) {
  for i := 0; i < 10; i++ {
    msg, err := dg.ChannelMessageSend(textChannelID, message)
    if err != nil {
      time.Sleep(1 * time.Second)
      continue
    }
    msgToPurgeQueue(msg)
    break
  }
}

func msgToPurgeQueue(m *discordgo.Message) {
  if purgeTime > 0 {
    timestamp := time.Now().UTC().Unix()
    message := PurgeMessage{
      m.ID,
      m.ChannelID,
      timestamp,
    }
    purgeQueue = append(purgeQueue, message)
  }
}

func purgeRoutine() { 
  if purgeTime > 0 {
    go func() {
      for {
        // TODO: There's even no need for range here, should be the zero element every time
        for k, v := range purgeQueue {
          if time.Now().Unix()-purgeTime > v.TimeSent {
            purgeQueue = append(purgeQueue[:k], purgeQueue[k+1:]...)
            dg.ChannelMessageDelete(v.ChannelID, v.ID)
            // Break at first match to avoid panic, timing isn't that important here
            break
          }
        }
        time.Sleep(time.Second * 1)
      }
    }()
  }
}

// GuildCreateHandler
func GuildCreateHandler(s *discordgo.Session, guild *discordgo.GuildCreate) {
  log.Println("INFO: Guild Create:", guild.ID)
  //GuildCreate(guild)
}

// GuildDeleteHandler
func GuildDeleteHandler(s *discordgo.Session, guild *discordgo.GuildDelete) {
  log.Println("INFO: Guild Delete:", guild.ID)
  v := voiceInstances[guild.ID]
  if v != nil {
    go func() {
      v.endSig <- true
    }()
    time.Sleep(200 * time.Millisecond)
    mutex.Lock()
    delete(voiceInstances, guild.ID)
    mutex.Unlock()
  }
}

// MessageCreateHandler
func MessageCreateHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
  if !strings.HasPrefix(m.Content, o.DiscordPrefix) {
    return
  }
  guildID := SearchGuild(m.ChannelID)
  v := voiceInstances[guildID]
  content := strings.Replace(m.Content, o.DiscordPrefix, "", 1)
  command := strings.Fields(content)
  switch(command[0]) {
    case "help", "h":
      HelpReporter(m)
    case "join", "j":
      JoinReporter(v, m)
    case "leave", "l":
      LeaveReporter(v, m)
    case "play":
      PlayReporter(v, m)
    case "radio":
      RadioReporter(v, m)
    case "stop":
      StopReporter(v, m)
    case "pause":
      PauseReporter(v, m)
    case "resume":
      ResumeReporter(v, m)
    case "queue":
      QueueReporter(v, m)
    case "skip":
      SkipReporter(v, m)
    case "youtube":
      YoutubeReporter(m)
    default:
      return
  }
}

