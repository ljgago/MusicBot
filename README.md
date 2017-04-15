# MusicBot

**MusicBot** is a multiserver music bot for Discord builded in Go. **MusicBot** plays youtube audio and url radio stream.

### Characteristics:
- Plays YouTube audio with query parameters or the url link.
- Plays url radio stream.
- Search YouTube videos.
- Support queue.
- Support remove song of queue by index, by user or by the last song.
- Support for skip, pause and resume.

### Build and install

You need to have installed in your system **go**, **ffmpeg** and **opus lib** (**opus** and **opusfile**)

```bash
# Install MusicBot
go get -u github.com/ljgago/MusicBot
```

### Use

**MusicBot** use a simple TOML config file.

```bash
MusicBot -f bot.toml
```

Example config file:

```bash
[discord]
  token = "YjQ4ODMyNTI0NzG3NDMwsDAw.CdNZBQ.fG5QVSUj7Gunf7CTTh69jG18tiQ" # Token bot
  status = "Music Bot | !!help"
  prefix = "!"

[youtube]
  token = "UIzRSyFyg75iDJbsKhaYk97UtgFriJjbo8uLH57"
```

License MIT.