package main

import (
  "errors"
  "log"
  "github.com/spf13/viper"
  "github.com/fsnotify/fsnotify"
)

type Options struct {
  DiscordToken    string
  YoutubeToken    string
  /*
  Guild           string
  Channel         string
  Token           string
  Status          string
  Url             string
  */
}

var o = &Options{}

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
  if o.YoutubeToken = viper.GetString("youtube.token"); o.YoutubeToken == "" {
    return errors.New("'token' must be present in config file")
  }


  /*
  if o.Guild = viper.GetString("discord.guild"); o.Guild == "" {
    return errors.New("'guild' must be present in config file")
  }
  if o.Channel = viper.GetString("discord.channel"); o.Channel == "" {
    return errors.New("'channel' must be present in config file")
  }
  
  if o.Status = viper.GetString("discord.status"); o.Status == "" {
    errors.New("'status' must be present in config file")
  }
  if o.Url = viper.GetString("discord.url"); o.Url == "" {
    errors.New("'url' must be present in config file")
  }
  */
  return nil
}

func Watch() {
  // Hot reload
  viper.WatchConfig()
  viper.OnConfigChange(Reload)
}

func Reload(e fsnotify.Event) {
  log.Println("INFO: The config file changed:", e.Name)
  LoadConfig(e.Name)
  //StopStream()
}