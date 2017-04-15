package main

import (
  "errors"
  "log"
  "github.com/spf13/viper"
  "github.com/fsnotify/fsnotify"
)

var o = &Options{}

// LoadConfig
func LoadConfig(filename string) (err error){
  // Read the config.toml file
  viper.SetConfigType("toml")
  viper.SetConfigFile(filename)
  //viper.AddConfigPath(".")
  err = viper.ReadInConfig()
  if err != nil {
    return err
  }
  
  if o.DiscordToken = viper.GetString("discord.token"); o.DiscordToken == "" {
    return errors.New("'token' must be present in config file")
  }
  if o.DiscordStatus = viper.GetString("discord.status"); o.DiscordStatus == "" {
    return errors.New("'status' must be present in config file")
  }
  if o.DiscordPrefix = viper.GetString("discord.prefix"); o.DiscordPrefix == "" {
    return errors.New("'prefix' must be present in config file")
  }
  if o.YoutubeToken = viper.GetString("youtube.token"); o.YoutubeToken == "" {
    return errors.New("'token' must be present in config file")
  }
  return nil
}

// Watch
func Watch() {
  // Hot reload
  viper.WatchConfig()
  viper.OnConfigChange(Reload)
}

// Reload
func Reload(e fsnotify.Event) {
  log.Println("INFO: The config file changed:", e.Name)
  LoadConfig(e.Name)
  //StopStream()
}