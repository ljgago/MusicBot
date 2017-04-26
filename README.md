# MusicBot

**MusicBot** is a multiserver music bot for Discord built in Go. **MusicBot** plays youtube audio and radio stream url.

### Features:
- Plays YouTube audio with query parameters or the url link.
- Plays radio stream url.
- Search YouTube videos.
- Support queue.
- Support remove song of queue by index, by user or by the last song.
- Support for skip, pause and resume.
- Support ignore commands of a channel.
- Support for message lifetime (config file)
- Support view title of song in status (config file) (Use this if you have one server only)
- Add Dockerfile and docker-compose for automatic build and run.

### Build and install

You need to have installed in your system **go>1.7** and **ffmpeg>3.0**

```bash
# Install MusicBot
go get -u github.com/ljgago/MusicBot
```

### Use

**MusicBot** use a simple TOML config file.

```bash
MusicBot -f bot.toml
```

### Docker

Edit and rename **_bot.toml.sample_** to **_bot.toml_**

```bash
# Run docker
docker build -t musicbot-img .
docker run -d --name musicbot --restart always -it musicbot-img
```

If you have docker-compose:

```bash
# Run docker-compose (automatic build and run)
docker-compose up -d
```

### Example bot.toml config file:

```bash
[discord]
  token = "YjQ4ODMyNTI0NzG3NDMwsDAw.CdNZBQ.fG5QVSUj7Gunf7CTTh69jG18tiQ" # Token bot
  status = "Music Bot | !help"
  prefix = "!"
  purgetime = 60 # message time to live 
  playstatus = false # Set 'true' if this bot run one server only

[youtube]
  token = "UIzRSyFyg75iDJbsKhaYk97UtgFriJjbo8uLH57"
```

License MIT.