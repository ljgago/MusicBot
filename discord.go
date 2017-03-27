package main

import (
  
  "log"
  "time"
  "strings"
  "github.com/bwmarrin/discordgo"
)



type Discord struct {
  Session     *discordgo.Session
}

// DiscordConnect make a new connection to Discord
func (d *Discord) DiscordConnect() (err error) {
  
  d.Session, err = discordgo.New("Bot " + o.Token)
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
  log.Println("INFO: Bot user test")
  log.Println("INFO: Playing", o.Status)
  log.Println("INFO:", o.Url)
  log.Println("INFO: Bot is now running. Press CTRL-C to exit.")
  return nil
}

// VoiceConnect join to voice channel and play audio streaming 
func (d *Discord) VoiceConnect() (err error) { 
  d.Session.UpdateStatus(0, o.Status)
  // Join to voice channel
  vs, err := d.Session.ChannelVoiceJoin(o.Guild, o.Channel, false, false)
  if err != nil {
    return err
  }
  PlayStream(vs, o.Url)
  return nil
}


// Start connect to discord and join to channel
func Start() (err error) {
  var d = &Discord{}
  if err = d.DiscordConnect(); err != nil {
    log.Println("FATA:", err)
    return err
  }
  for {
    if err = d.VoiceConnect(); err != nil {
      log.Println("FATA:", err)
    }
    log.Println("INFO: Restarting ...")
    time.Sleep(5000 * time.Millisecond)
  }
  return nil
}


// messageCreate handler for controller text input
func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
  if len(m.Mentions) == 0 {
    return
  }
  // Restart the bot
  botID, _ := s.User("@me")
  if m.Mentions[0].ID != botID.ID {
    return
  }

  if strings.Contains(m.Content, "!play") {
    log.Println("INFO:", m.Author.Username, "send '!play'")
    //playReporter(s, m)
    return
  }
}