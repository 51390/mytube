package mytube

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
    "os"
	"strings"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/youtube/v3"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func getClientWithToken(ctx context.Context, config *oauth2.Config, tokenType string, token string) *http.Client {
	t := &oauth2.Token{}
    err := json.Unmarshal([]byte(token), t)
    if err != nil {
        log.Fatalf("Unable to parse token: %s", token)
    }
    return config.Client(ctx, t)
}

func HandleError(err error, message string) {
	if message == "" {
		message = "Error making API call"
	}
	if err != nil {
		log.Fatalf(message+": %v", err.Error())
	}
}

func channelsListById(ctx context.Context, service *youtube.Service, id ...string) *[]string{
	perCall := 50
	numCalls := len(id) / perCall
	playlistIds := []string{}

	for i := 0; i < numCalls; i++ {
		call := service.Channels.List([]string{"snippet", "contentDetails", "statistics"})
		call = call.Id(id[i*perCall : (i+1)*perCall]...)
		call.Pages(ctx, func(response *youtube.ChannelListResponse) error {
			for _, channel := range response.Items {
				relatedPlaylists := channel.ContentDetails.RelatedPlaylists
				playlistIds = append(playlistIds, relatedPlaylists.Uploads, relatedPlaylists.Likes, relatedPlaylists.Favorites)
			}

            log.Printf("%d playlists loaded.", len(playlistIds))
			return nil
		})
	}

    return &playlistIds
}

func subscriptionsList(ctx context.Context, service *youtube.Service) *[]string {
	subscriptionService := youtube.NewSubscriptionsService(service)
	call := subscriptionService.List([]string{"snippet", "contentDetails"})
	numChannels := 0
	channelIds := []string{}
    err := call.Mine(true).Pages(ctx, func(response *youtube.SubscriptionListResponse) error {
		for _, item := range response.Items {
			numChannels += 1
			channelId := item.Snippet.ResourceId.ChannelId
			channelIds = append(channelIds, channelId)
		}
        log.Printf("%d channels loaded.", len(channelIds))
		return nil
	})

    if err != nil {
        log.Printf("Failed loading channel subscriptions: %s\n", err)
    }

    return &channelIds
}

func playlistItemsByPlaylistId(ctx context.Context, service *youtube.Service, playlistIds ...string) *[]string {
	playlistsItemService := youtube.NewPlaylistItemsService(service)
	videoIds := []string{}
	for _, id := range playlistIds {
		call := playlistsItemService.List([]string{"snippet", "contentDetails"}).PlaylistId(id)
		call.Pages(ctx, func(response *youtube.PlaylistItemListResponse) error {
			for _, item := range response.Items {
				videoIds = append(videoIds, item.ContentDetails.VideoId)
			}
            log.Printf("%d videos loaded.", len(videoIds))

			return errors.New("first page loaded")
		})
	}

    return &videoIds
}

func thumbnailUrl(video *youtube.Video) string {
    if video.Snippet.Thumbnails.Medium != nil {
        return video.Snippet.Thumbnails.Medium.Url
    } else if video.Snippet.Thumbnails.Default != nil {
        return video.Snippet.Thumbnails.Default.Url
    } else {
        return ""
    }
}

func videoDetails(ctx context.Context, service *youtube.Service, ids []string) {
	perCall := 50
	numCalls := len(ids) / perCall
	videosService := youtube.NewVideosService(service)
	videos := []*Video{}
    userId := ctx.Value("userId").(string)

	for i := 0; i < numCalls; i++ {
		idBatch := ids[i*perCall : (i+1)*perCall]
		videosListCall := videosService.List([]string{"snippet", "contentDetails", "statistics", "player"}).Id(idBatch...)
		videosListCall.Pages(ctx, func(response *youtube.VideoListResponse) error {
			for _, item := range response.Items {
				video_item := NewVideo(
					userId, item.Id, item.Snippet.Title, item.Snippet.Description,
					item.Snippet.ChannelId, item.Snippet.ChannelTitle,
					item.Statistics.ViewCount, item.Statistics.LikeCount,
					item.Statistics.CommentCount, item.Statistics.DislikeCount,
					item.Statistics.FavoriteCount, item.ContentDetails.Duration,
					item.Snippet.PublishedAt, strings.Join(item.Snippet.Tags, ","),
                    thumbnailUrl(item),
				)
				videos = append(videos, &video_item)
				db, ok := ctx.Value("db").(*gorm.DB)
				if ok {
                    result := db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&video_item)
                    if result.Error != nil {
                        log.Printf("Error saving %s / %s: %s\n", video_item.UserId, video_item.Id, result.Error)
                    } else {
                        log.Printf("Saved video id %s / %s\n", video_item.UserId, video_item.Id)
                    }
				} else {
                    log.Printf("Unable to get db connection from context, skipping.")
                }
			}
			return nil
		})
	}
}

func configToJson() string {
    json := `{

        "installed": {
            "client_id": "%s",
            "client_secret": "%s",
            "project_id": "%s",
            "auth_uri": "%s",
            "token_uri": "%s",
            "auth_provider_x509_cert_url": "%s",
            "redirect_uris": %s
        }
    }`

    return fmt.Sprintf(json, os.Getenv("CLIENT_ID"), os.Getenv("CLIENT_SECRET"), os.Getenv("PROJECT_ID"), os.Getenv("AUTH_URI"), os.Getenv("TOKEN_URI"), os.Getenv("AUTH_PROVIDER_X509_CERT_URL"), os.Getenv("REDIRECT_URIS"))
}

func Sync(userId string, tokenType string, token string) {
	ctx := context.Background()
    ctx = context.WithValue(ctx, "userId", userId)

	db, err := InitDb()
	HandleError(err, "Failed initializing db")
	Migrate(db)
    ctx = context.WithValue(ctx, "db", db)

    configJson := configToJson()
	config, err := google.ConfigFromJSON([]byte(configJson), youtube.YoutubeReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

    client :=  getClientWithToken(ctx, config, tokenType, token)
	service, err := youtube.New(client)

	HandleError(err, "Error creating YouTube client")

    channelIds := subscriptionsList(ctx, service)
    playlistIds := channelsListById(ctx, service, *channelIds...)
    videoIds := playlistItemsByPlaylistId(ctx, service, *playlistIds...)
	videoDetails(ctx, service, *videoIds)
}

func max(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}
