package yourtube

import (
  "errors"
  "encoding/json"
  "fmt"
  "log"
  "io/ioutil"
  "net/http"
  "net/url"
  "os"
  "os/user"
  "path/filepath"
  "strings"

  "golang.org/x/net/context"
  "golang.org/x/oauth2"
  "golang.org/x/oauth2/google"
  "google.golang.org/api/youtube/v3"
  "gorm.io/gorm"
  "gorm.io/gorm/clause"
)

const missingClientSecretsMessage = `
Please configure OAuth 2.0
`

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
  cacheFile, err := tokenCacheFile()
  if err != nil {
    log.Fatalf("Unable to get path to cached credential file. %v", err)
  }
  tok, err := tokenFromFile(cacheFile)
  if err != nil {
    tok = getTokenFromWeb(config)
    saveToken(cacheFile, tok)
  }
  return config.Client(ctx, tok)
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
  authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
  fmt.Printf("Go to the following link in your browser then type the "+
    "authorization code: \n%v\n", authURL)

  var code string
  if _, err := fmt.Scan(&code); err != nil {
    log.Fatalf("Unable to read authorization code %v", err)
  }

  tok, err := config.Exchange(oauth2.NoContext, code)
  if err != nil {
    log.Fatalf("Unable to retrieve token from web %v", err)
  }
  return tok
}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile() (string, error) {
  usr, err := user.Current()
  if err != nil {
    return "", err
  }
  tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
  os.MkdirAll(tokenCacheDir, 0700)
  return filepath.Join(tokenCacheDir,
    url.QueryEscape("youtube-go-quickstart.json")), err
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
  f, err := os.Open(file)
  if err != nil {
    return nil, err
  }
  t := &oauth2.Token{}
  err = json.NewDecoder(f).Decode(t)
  defer f.Close()
  return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) {
  fmt.Printf("Saving credential file to: %s\n", file)
  f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
  if err != nil {
    log.Fatalf("Unable to cache oauth token: %v", err)
  }
  defer f.Close()
  json.NewEncoder(f).Encode(token)
}

func HandleError(err error, message string) {
  if message == "" {
    message = "Error making API call"
  }
  if err != nil {
    log.Fatalf(message + ": %v", err.Error())
  }
}

func channelsListById(ctx context.Context, service *youtube.Service, id ...string) {
  perCall := 50
  numCalls := len(id) / perCall
  playlistIds := []string{}

  for i := 0; i < numCalls; i++ {
      call := service.Channels.List([]string{"snippet", "contentDetails", "statistics"})
      call = call.Id(id[i * perCall : (i+1) * perCall]...)
      call.Pages(ctx, func (response *youtube.ChannelListResponse) (error) {
          for _, channel := range response.Items {
              relatedPlaylists := channel.ContentDetails.RelatedPlaylists

              fmt.Println(fmt.Sprintf("\tThis channel's ID is %s. Its title is '%s', " +
              "and it has %d views.Relevant playlists: Uploads: %s | Likes: %s | Favorites: %s",
              channel.Id,
              channel.Snippet.Title,
              channel.Statistics.ViewCount,
              relatedPlaylists.Uploads,
              relatedPlaylists.Likes,
              relatedPlaylists.Favorites))

             playlistIds = append(playlistIds, relatedPlaylists.Uploads, relatedPlaylists.Likes, relatedPlaylists.Favorites)
          }

          return nil
      })
  }

  playlistItemsByPlaylistId(ctx, service, playlistIds...)
}

func subscriptionsList(ctx context.Context, service *youtube.Service) {
    subscriptionService := youtube.NewSubscriptionsService(service)
    call := subscriptionService.List([]string{"snippet", "contentDetails"})
    numChannels := 0
    channelIds := []string{}
    call.Mine(true).Pages(ctx, func (response *youtube.SubscriptionListResponse) (error) {
        for _, item := range response.Items {
            numChannels += 1
            channelId := item.Snippet.ResourceId.ChannelId
            fmt.Printf("%d: %s / %s / %s\n", numChannels, item.Snippet.Title, channelId, item.Id)
            channelIds = append(channelIds, channelId)
        }
        return nil
    })

    fmt.Println("Done")

    channelsListById(ctx, service, channelIds...)
}

func playlistItemsByPlaylistId(ctx context.Context, service *youtube.Service, playlistIds ...string) {
    playlistsItemService := youtube.NewPlaylistItemsService(service)
    videoIds := []string{}
    for _, id := range playlistIds {
        call := playlistsItemService.List([]string{"snippet", "contentDetails"}).PlaylistId(id)
        call.Pages(ctx, func(response *youtube.PlaylistItemListResponse) (error) {
            for _, item := range response.Items {
                videoIds = append(videoIds, item.ContentDetails.VideoId)
            }

            return errors.New("first page loaded")
        })
    }

    videoDetails(ctx, service, videoIds)
}

func videoDetails(ctx context.Context, service *youtube.Service, ids []string) []*video {
    perCall := 50
    numCalls := len(ids) / perCall
    videosService := youtube.NewVideosService(service)
    videos := []*video{};

    for i := 0; i < numCalls; i++ {
        idBatch := ids[i * perCall : (i+1) * perCall]
        videosListCall := videosService.List([]string{"snippet", "contentDetails", "statistics"}).Id(idBatch...)
        videosListCall.Pages(ctx, func(response *youtube.VideoListResponse) (error) {
            for _, item := range response.Items {
                fmt.Printf("\t V>> %s %s\n", item.Snippet.Title, item.ContentDetails.Duration)
                video_item := NewVideo(
                    item.Id, item.Snippet.Title, item.Snippet.Description,
                    item.Snippet.ChannelId, item.Snippet.ChannelTitle,
                    item.Statistics.ViewCount, item.Statistics.LikeCount,
                    item.Statistics.CommentCount, item.Statistics.DislikeCount,
                    item.Statistics.FavoriteCount, item.ContentDetails.Duration,
                    item.Snippet.PublishedAt, strings.Join(item.Snippet.Tags, ","),
                )
                videos = append(videos, &video_item)
                db, ok := ctx.Value("db").(*gorm.DB)
                if ok {
                    db.Clauses(clause.OnConflict{ UpdateAll:true }).Create(&video_item)
                }
            }
            return nil
        })
    }

    return videos
}

func Sync(db *gorm.DB) {
  ctx := context.Background()

  b, err := ioutil.ReadFile("client_secret.json")
  if err != nil {
    log.Fatalf("Unable to read client secret file: %v", err)
  }

  // If modifying these scopes, delete your previously saved credentials
  // at ~/.credentials/youtube-go-quickstart.json
  config, err := google.ConfigFromJSON(b, youtube.YoutubeReadonlyScope)
  if err != nil {
    log.Fatalf("Unable to parse client secret file to config: %v", err)
  }
  client := getClient(ctx, config)
  service, err := youtube.New(client)

  HandleError(err, "Error creating YouTube client")

  subscriptionsList(context.WithValue(ctx, "db", db), service);
}

func max(a, b int) int {
    if a > b {
        return a
    } else {
        return b
    }
}



