package main

import (
  //"time"
  "regexp"
  "net/http"
  //"strconv"
  "github.com/google/google-api-go-client/googleapi/transport"
  "google.golang.org/api/youtube/v3"
  "github.com/rylio/ytdl"
  "github.com/bwmarrin/discordgo"
)


func YoutubeFind(searchString string, m *discordgo.MessageCreate) (song Song, err error) { //(url, title, time string, err error)
  // YouTube
  var rxpDurationDays, rxpDurationHours, rxpDurationMinutes, rxpDurationSeconds *regexp.Regexp

  rxpDurationDays = regexp.MustCompile(`([0-9]*)D`)
  rxpDurationHours = regexp.MustCompile(`([0-9]*)H`)
  rxpDurationMinutes = regexp.MustCompile(`([0-9]*)M`)
  rxpDurationSeconds = regexp.MustCompile(`([0-9]*)S`)

  client := &http.Client{
    Transport: &transport.APIKey{Key: o.YoutubeToken},
  }

  service, err := youtube.New(client)
  if err != nil {
    //log.Fatalf("Error creating new YouTube client: %v", err)
    return
  }

  call := service.Search.List("id,snippet").Q(searchString).MaxResults(1)
  response, err := call.Do()
  if err != nil {
    //log.Fatalf("Error making search API call: %v", err)
    return
  }

  var (
    audioId, audioTitle string //, fileVideoID string
  )

  for _, item := range response.Items {
    audioId = item.Id.VideoId
    audioTitle = item.Snippet.Title
  }
  if audioId == "" {
    ChMessageSend(m.ChannelID, "Sorry, I can't found this song.")
    return
  }
  vid, err := ytdl.GetVideoInfo("https://www.youtube.com/watch?v=" + audioId)
  if err != nil {
    //ChMessageSend(textChannelID, "Sorry, nothing found for query: "+strings.Trim(searchString, " "))
    return
  }
  format := vid.Formats.Extremes(ytdl.FormatAudioBitrateKey, true)[0]
  videoURL, _ := vid.GetDownloadURL(format)
  videoURLString := videoURL.String()

  videos := service.Videos.List("contentDetails").Id(vid.ID)
  resp, err := videos.Do()
  
  var (
    duration, durationString string
  )

  duration = resp.Items[0].ContentDetails.Duration

  // TODO: Rewrite this parsing bit
  if rxpDurationDays.FindStringSubmatch(duration) != nil {
    durationString = durationString + rxpDurationDays.FindStringSubmatch(duration)[1] + ":"
  }

  if rxpDurationHours.FindStringSubmatch(duration) != nil {
    if rxpDurationDays.FindStringSubmatch(duration) != nil {
      if len(rxpDurationHours.FindStringSubmatch(duration)[1]) == 1 {
        durationString = durationString + "0" + rxpDurationHours.FindStringSubmatch(duration)[1] + ":"
      } else {
        durationString = durationString + rxpDurationHours.FindStringSubmatch(duration)[1] + ":"
      }
    } else {
      durationString = durationString + rxpDurationHours.FindStringSubmatch(duration)[1] + ":"
    }
  }

  if rxpDurationMinutes.FindStringSubmatch(duration) != nil {
    if rxpDurationHours.FindStringSubmatch(duration) != nil {
      if len(rxpDurationMinutes.FindStringSubmatch(duration)[1]) == 1 {
        durationString = durationString + "0" + rxpDurationMinutes.FindStringSubmatch(duration)[1] + ":"
      } else {
        durationString = durationString + rxpDurationMinutes.FindStringSubmatch(duration)[1] + ":"
      }
    } else {
      durationString = durationString + rxpDurationMinutes.FindStringSubmatch(duration)[1] + ":"
    }
  } else {
    durationString = durationString + "00:"
  }

  if rxpDurationSeconds.FindStringSubmatch(duration) != nil {
    if len(rxpDurationSeconds.FindStringSubmatch(duration)[1]) == 1 {
      durationString = durationString + "0" + rxpDurationSeconds.FindStringSubmatch(duration)[1]
    } else {
      durationString = durationString + rxpDurationSeconds.FindStringSubmatch(duration)[1]
    }
  } else {
    durationString = durationString + "00"
  }

  song = Song{
    m.ChannelID,
    m.Author.Username,
    vid.ID,
    audioTitle,
    durationString,
    videoURLString,
  }
  //vi.queue = append(queue, song)
  return 
}
