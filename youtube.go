package main

import (
  //"log"
  "net/http"
	"github.com/google/google-api-go-client/googleapi/transport"
	"google.golang.org/api/youtube/v3"
  "github.com/rylio/ytdl"
)

type Song struct {
  ID          string
  QueueID     string
  Title       string
  OrderedBy   string
  Duration    string
  Status      string
  VideoURL    string
  VK          bool
}

var (
  queue       []Song
  //queueLock = &sync.Mutex{}

)

func youtubeFind(searchString string) (url, title string, err error) {
  //queueLock.Lock()
  //defer queueLock.Unlock()

  // YouTube
  client := &http.Client{
    Transport: &transport.APIKey{Key: o.YoutubeToken},
  }

  service, err := youtube.New(client)
  if err != nil {
    //log.Fatalf("Error creating new YouTube client: %v", err)
    return "", "", err
  }

  call := service.Search.List("id,snippet").Q(searchString).MaxResults(1)
  response, err := call.Do()
  if err != nil {
    //log.Fatalf("Error making search API call: %v", err)
    return "", "", err
  }

  var (
    audioId, audioTitle string//fileVideoID string
  )

  for _, item := range response.Items {
    audioId = item.Id.VideoId
    audioTitle = item.Snippet.Title
  }

  vid, err := ytdl.GetVideoInfo("https://www.youtube.com/watch?v=" + audioId)
  if err != nil {
    //ChMessageSend(textChannelID, "Sorry, nothing found for query: "+strings.Trim(searchString, " "))
    return "", "", err
  }
  format := vid.Formats.Extremes(ytdl.FormatAudioBitrateKey, true)[0]
  videoURL, _ := vid.GetDownloadURL(format)
  videoURLString := videoURL.String()

  return videoURLString, audioTitle, nil

  /*
  if len(queue) > 0 {
    for _, v := range queue {
      if v.ID == vid.ID {
        fileVideoID = vid.ID + "_" + strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
        break
      } else {
        fileVideoID = vid.ID
      }
    }
  } else {
    fileVideoID = vid.ID
  }

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

  song := Song{
    vid.ID,
    fileVideoID,
    audioTitle,
    m.Author.Username,
    durationString,
    "q",
    videoURLString,
    false,
  }

  queue = append(queue, song)

  if forRandom == false {
    ChMessageSend(textChannelID, "Enqueued **"+audioTitle+"** ["+durationString+"]")
  }

  go func() {
    status <- song
  }()
  */
}
