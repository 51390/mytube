package yourtube

import (
	"fmt"
	"github.com/sosodev/duration"
	"gorm.io/gorm"
)

type Video struct {
	gorm.Model

	Id            string `gorm:"primaryKey"`
	ChannelId     string
	ChannelTitle  string
	CommentCount  uint64
	Description   string
	DislikeCount  uint64
	Duration      string
	FavoriteCount uint64
	LikeCount     uint64
	PublishedAt   string
	Tags          string
	Title         string
	ViewCount     uint64
    Thumbnail     string
}

func NewVideo(id string, title string, description string,
	channelId string, channelTitle string,
	viewCount uint64, likeCount uint64, commentCount uint64,
	dislikeCount uint64, favoriteCount uint64, durationString string,
	publishedAt string, tags string, thumbnail string) Video {

	d, err := duration.Parse(durationString)
	if err == nil {
		durationString = fmt.Sprintf("%02d:%02d:%02d", uint64(d.Hours), uint64(d.Minutes), uint64(d.Seconds))
	}
	return Video{
		Id:            id,
		Title:         title,
		Description:   description,
		ChannelId:     channelId,
		ChannelTitle:  channelTitle,
		CommentCount:  commentCount,
		DislikeCount:  dislikeCount,
		LikeCount:     likeCount,
		FavoriteCount: favoriteCount,
		PublishedAt:   publishedAt,
		ViewCount:     viewCount,
		Duration:      durationString,
        Thumbnail:     thumbnail,
	}
}
