package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	//"github.com/go-chi/httplog"
	"github.com/go-chi/render"
	"gorm.io/gorm"

	"yourtube.51390.cloud/yourtube"
)

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

func (vr *VideoResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func NewVideoResponse(video *yourtube.Video) *VideoResponse {
	return &VideoResponse{Video: video}
}

func NewVideoListResponse(videos *[]yourtube.Video) (response []render.Renderer) {
	for _, video := range *videos {
		response = append(response, NewVideoResponse(&video))
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

	fmt.Println(clause)

	return db.Where(clause)
}

func videos(w http.ResponseWriter, r *http.Request) {
	//log := httplog.LogEntry(r.Context())
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

	for param, value := range r.Form {
		fmt.Printf("--> %s / %s\n", param, value)
		db = parseFilter(db, param, value[0])
	}

	fmt.Println("Test")

	videos := []yourtube.Video{}
	if err = db.Where("").Find(&videos).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else if err = render.RenderList(w, r, NewVideoListResponse(&videos)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func main() {
	// http
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(dbMiddleware)
	r.Get("/videos", videos)
	http.ListenAndServe(":3000", r)
}
