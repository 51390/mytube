package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"gorm.io/gorm"

	"yourtube.51390.cloud/yourtube"
)

func contentTypeMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
    })
}

func dbMiddleware(next http.Handler) http.Handler {
	db, err := yourtube.InitDb()
	yourtube.HandleError(err, "Failed initializing database")
	yourtube.Migrate(db)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timeoutContext, _ := context.WithTimeout(
			context.Background(), time.Second*5)
		ctx := context.WithValue(
			r.Context(), "db", db.WithContext(timeoutContext))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func db(r *http.Request) (db *gorm.DB, err error) {
	db, ok := r.Context().Value("db").(*gorm.DB)
	if !ok {
		err = errors.New("Failed loading DB from request context")
	}

	return
}

type VideoResponse struct {
	*yourtube.Video
}

func (vr VideoResponse) MarshalJSON() ([]byte, error) {
	fields := make(map[string]interface{})
	fields["id"] = vr.Id
	fields["channelId"] = vr.ChannelId
	fields["channelTitle"] = vr.ChannelTitle
	fields["commentCount"] = vr.CommentCount
	fields["description"] = vr.Description
	fields["dislikeCount"] = vr.DislikeCount
	fields["duration"] = vr.Duration
	fields["favoriteCount"] = vr.FavoriteCount
	fields["likeCount"] = vr.LikeCount
	fields["publishedAt"] = vr.PublishedAt
	fields["tags"] = vr.Tags
	fields["title"] = vr.Title
	fields["viewCount"] = vr.ViewCount
    fields["publishedAt"] = vr.PublishedAt
    fields["thumbnail"] = vr.Thumbnail

	return json.Marshal(fields)
}

func (vr *VideoResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func NewVideoResponse(video *yourtube.Video) *VideoResponse {
	return &VideoResponse{Video: video}
}

func NewVideoListResponse(videos []*yourtube.Video) (response []render.Renderer) {
	for _, video := range videos {
        response = append(response, NewVideoResponse(video))
	}

	return
}

func parseFilter(db *gorm.DB, filter string, value string) *gorm.DB {
	pattern := regexp.MustCompile("(\\w+)\\[([^\\s]+)\\]")
	matches := pattern.FindStringSubmatch(filter)

	var clause string

	if len(matches) >= 3 {
		field := matches[1]
		operation := matches[2]
		clause = fmt.Sprintf("%s %s %s", field, operation, value)
	} else {
		clause = fmt.Sprintf("%s = %s", filter, value)
	}

	return db.Where(clause)
}

func videos(w http.ResponseWriter, r *http.Request) {
    userId := chi.URLParam(r, "userId")
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	db, err := db(r)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

    db = db.Where("user_id = ?", userId)

	for param, value := range r.Form {
		db = parseFilter(db, param, value[0])
	}

	videos := []*yourtube.Video{}
	if err = db.Find(&videos).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else if err = render.RenderList(w, r, NewVideoListResponse(videos)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func syncVideos(w http.ResponseWriter, r *http.Request) {
    userId := chi.URLParam(r, "userId")
    token := r.Header.Get("Authorization")
    tokenType := "Bearer"
    fmt.Println("Syncing videos for", userId)
    go yourtube.Sync(userId, tokenType, token)
}

func main() {
	// http
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(dbMiddleware)
	r.Get("/videos/{userId}", videos)
    r.Post("/videos/sync/{userId}", syncVideos)
	http.ListenAndServe(":3000", r)
}
