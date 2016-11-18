# musicbot

**musicbot** is a simple music bot for Discord builded in Go. **musicbot** play url stream and single files. It use [discordgo](https://github.com/bwmarrin/discordgo), [dgvoice](https://github.com/bwmarrin/dgvoice), and [viper](https://github.com/spf13/viper) for config file and hot reload.

### Build and install

You need to have installed in your system _go_ and for _dgvoice_ you need:

* You must use the current develop branch of Discordgo
* You must have _ffmpeg_ in your path and _Opus libs_ already installed.

```bash
# Install discordgo
go get -u github.com/bwmarrin/discordgo
# Change to develop branch
cd $GOPATH/src/github.com/bwmarrin/discordgo
git checkout develop
# Rebuild
go install github.com/bwmarrin/discordgo
# Install dgvoice
go get -u github.com/bwmarrin/dgvoice
# Install musicbot
go get -u github.com/ljgago/musicbot
```

### Use

**musicbot** use a simple TOML config file.

```bash
musicbot -f bot.toml
```

Example config file:

```bash
[bot]
  guild_id = "724349134172233488" # The server id
  channel_id = "208643566488517230" # Voice channel id
  token = "fjQ4ODfydTI0efA3NDgwNDAw.Cw98dQ.GETgVfjrMh6fCp6GH34EcdvnRvI" # Token bot
  status = "Music" # Status bot
  url = "http://audio.misproductions.com/japan48k" # Url streaming
```

If the bot is running and the config file is modified, **musicbot** reload automatic the new config (hot reload).

License MIT.