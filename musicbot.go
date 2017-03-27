package main

import (
  
  "log"
  "flag"
)

func main() {
  filename := flag.String("f", "bot.toml", "Set path for the config file.")
  flag.Parse()
  
  log.Println("INFO: Opening", *filename)
  err := LoadConfig(*filename)
  if err != nil {
    log.Println("FATA:", err)
    return
  }
  Watch()
  err = Start()
  if err != nil {
    return
  }
  <-make(chan struct{})
}

/*
// Hot reload
  viper.WatchConfig()
  viper.OnConfigChange(func (e fsnotify.Event) {
    log.Println("The config file changed:", e.Name)
    KillPlayer()
  })

  */
