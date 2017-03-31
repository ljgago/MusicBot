package main

import (
  
  "log"
  "time"
  "errors"
  "strings"
  "github.com/bwmarrin/discordgo"
)

type Discord struct {
  Session     *discordgo.Session
}

// DiscordConnect make a new connection to Discord
func (d *Discord) DiscordConnect() (err error) {
  
  d.Session, err = discordgo.New("Bot " + o.DiscordToken)
  if err != nil {
    log.Println("FATA: error creating Discord session,", err)
    return
  }
  log.Println("INFO: Bot is Opening")
  d.Session.AddHandler(messageHandler)
  // Open Websocket
  err = d.Session.Open()
  if err != nil {
    log.Println("FATA: Error Open():", err)
    return
  }
  _, err = d.Session.User("@me")
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
  return nil
}

// VoiceConnect join to voice channel and play audio streaming 
/*
func (d *Discord) VoiceConnect() (err error) { 
  //d.Session.UpdateStatus(0, o.Status)
  // Join to voice channel
  vs, err := d.Session.ChannelVoiceJoin(o.Guild, o.Channel, false, false)
  if err != nil {
    return err
  }
  PlayStream(vs, o.Url)
  return nil
}
*/

// Start connect to discord and join to channel
func Start() (err error) {
  var d = &Discord{}
  if err = d.DiscordConnect(); err != nil {
    log.Println("FATA:", err)
    return err
  }
  /*
  for {
    if err = d.VoiceConnect(); err != nil {
      log.Println("FATA:", err)
    }
    log.Println("INFO: Restarting ...")
    time.Sleep(5000 * time.Millisecond)
  }
  */
  return nil
}

// searchVoiceChannel search the voice channel id into from guild.
func searchVoiceChannel(s *discordgo.Session, m *discordgo.MessageCreate) (channelID string, err error) { 
  for _, g := range s.State.Guilds {
    for _, v := range g.VoiceStates {
      if v.UserID == m.Author.ID {
        return v.ChannelID, nil
      }
    }   
  }
  return "", errors.New("The user is not in a voice channel.")
}

func searchGuild(s *discordgo.Session, m *discordgo.MessageCreate) (guildID string) {
  channel, _ := s.Channel(m.ChannelID)
  return channel.GuildID
}

func getVoiceConnection(s *discordgo.Session, guild string) (voice *discordgo.VoiceConnection) {
  return (s.VoiceConnections[guild])
}


func retryOnBadGateway(f func() error) {
  var err error
  for i := 0; i < 3; i++ {
    if err = f(); err != nil {
      if strings.HasPrefix(err.Error(), "HTTP 502") {
        // If the error is Bad Gateway, try again after 1 sec.
        time.Sleep(1 * time.Second)
        continue
      } else {
        // Otherwise panic !
        log.Println("ERROR: ", err)
      }
    } else {
      // In case of no error, return.
      return
    }
  }
}

func sendMessage(s *discordgo.Session, m *discordgo.MessageCreate, msg string) (err error) {
  //log.Println("BOT: ", msg)
  retryOnBadGateway(func() error {
    return sendFormattedMessage(s, m, msg)
  })
  return
}

func sendFormattedMessage(s *discordgo.Session, m *discordgo.MessageCreate, msg string) (err error) {
  prefix := ":notes: | **" + m.Author.Username + "**, "
  _, err = s.ChannelMessageSend(m.ChannelID, prefix + msg)
  if err != nil {
    return err
  }
  return nil
}

// messageCreate handler for controller text input
func (d *Discord) messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
  if len(m.Mentions) == 0 {
    return
  }
  // Restart the bot
  botID, _ := s.User("@me")
  if m.Mentions[0].ID != botID.ID {
    return
  }

  if strings.Contains(m.Content, "!help") {
    log.Println("INFO:", m.Author.Username, "send '!help'")
    helpReporter(s, m)
    return
  }

  if strings.Contains(m.Content, "!play") {
    log.Println("INFO:", m.Author.Username, "send '!play'")
    playReporter(s, m)
    return
  }
  if strings.Contains(m.Content, "!stop") {
    log.Println("INFO:", m.Author.Username, "send '!stop'")
    stopReporter(s, m)
    return
  }
  if strings.Contains(m.Content, "!youtube") {
    log.Println("INFO:", m.Author.Username, "send '!youtube'")
    youtubeReporter(s, m)
    return
  }
  if strings.Contains(m.Content, "!status") {
    log.Println("INFO:", m.Author.Username, "send '!status'")
    statusReporter(s, m)
    return
  }
  if strings.Contains(m.Content, "!clean") {
    log.Println("INFO:", m.Author.Username, "send '!clean'")
    cleanReporter(s, m)
    return
  }
}

func helpReporter(s *discordgo.Session, m *discordgo.MessageCreate) {
  help := "```go\n`Standard Commands List`\n```\n" + 
  "`!help` -> show help commands.\n" +
  "`!play` -> play the url streaming. Use !play url.\n" +
  "`!stop` -> stop the played streaming.\n" + 
  "`!status` -> change the status of the bot.\n" + 
  "`!clean` -> clean the status of the bot.\n"

  sendMessage(s, m, help)
}

func playReporter(s *discordgo.Session, m *discordgo.MessageCreate) {
  guildID := searchGuild(s, m)
  channelID, err := searchVoiceChannel(s, m)
  if err != nil {
    log.Println("ERROR: Voice channel id not found: ", err)
    sendMessage(s, m, "you need to be in a voice channel.")
    return
  }

  command := strings.Fields(m.Content)

  if len(command) != 3 {
    sendMessage(s, m, "you need to specify an url.")
    return
  }
  url := command[2]
  vi := voiceInstances[guildID]
  if vi != nil {
    // change audio to the voice channel
    vi.StopStream()
    log.Println("INFO: Update voice channel")
    time.Sleep(5000 * time.Millisecond)
  }

  // create new voice instance
  vi = new(VoiceInstance)
  voiceInstances[guildID] = vi
  vi.guildID = guildID
  
  if vi.voice, err = s.ChannelVoiceJoin(guildID, channelID, false, false); err != nil {
    delete(voiceInstances, vi.guildID)
    log.Println("ERROR: Error to join in a voice channel: ", err)
    return
  }
  log.Println("INFO: New voice channel created")

  // Play to URL
  vi.end = make(chan bool)
  go vi.PlayStream(url, vi.end)
  sendMessage(s, m, "I'm playing now!")

}

func stopReporter(s *discordgo.Session, m *discordgo.MessageCreate) {
  //s.UpdateStatus(0, "")
  guildID := searchGuild(s, m)
  vi := voiceInstances[guildID] 
  if vi == nil {
    log.Println("INFO: The bot get ready stoped")
    sendMessage(s, m, "I'm ready stoped!")
    return
  }
  vi.StopStream()
  //delete(voiceInstances, guildID)
  //voiceInstances[guildID] = nil
  log.Println("INFO: The bot stop play audio")
  sendMessage(s, m, "I'm stop now!")
}

func youtubeReporter(s *discordgo.Session, m *discordgo.MessageCreate) {
  guildID := searchGuild(s, m)
  channelID, err := searchVoiceChannel(s, m)
  if err != nil {
    log.Println("ERROR: Voice channel id not found: ", err)
    sendMessage(s, m, "you need to be in a voice channel.")
    return
  }

  if len(strings.Fields(m.Content)) < 3 {
    sendMessage(s, m, "you need to specify a name.")
    return
  }
  command := strings.SplitAfter(m.Content, "!youtube")
  queue := strings.TrimSpace(command[1])

  url, title, err := youtubeFind(queue)
  if err != nil {
    log.Println("ERROR: Youtube search: ", err)
  }

  vi := voiceInstances[guildID]
  if vi != nil {
    // change audio to the voice channel
    vi.StopStream()
    log.Println("INFO: Update voice channel")
    time.Sleep(5000 * time.Millisecond)
  }

  // create new voice instance
  vi = new(VoiceInstance)
  voiceInstances[guildID] = vi
  vi.guildID = guildID
  
  if vi.voice, err = s.ChannelVoiceJoin(guildID, channelID, false, false); err != nil {
    delete(voiceInstances, vi.guildID)
    log.Println("ERROR: Error to join in a voice channel: ", err)
    return
  }
  log.Println("INFO: New voice channel created")

  // Play to URL
  vi.end = make(chan bool)
  go vi.PlayStream(url, vi.end)
  sendMessage(s, m, "I'm playing `" + title + "`")
}

func statusReporter(s *discordgo.Session, m *discordgo.MessageCreate) {
  if len(strings.Fields(m.Content)) < 3 {
    sendMessage(s, m, "you need to specify a status.")
    return
  }
  command := strings.SplitAfter(m.Content, "!status")
  status := strings.TrimSpace(command[1])
  s.UpdateStatus(0, status)
  sendMessage(s, m, "I'm playing `" + status + "`")
}

func cleanReporter(s *discordgo.Session, m *discordgo.MessageCreate) {
  s.UpdateStatus(0, "")
}

