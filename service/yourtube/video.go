package yourtube

import (
    "gorm.io/gorm"
)

type video struct {
    gorm.Model

    Id string
    ChannelId string
    ChannelTitle string
    CommentCount uint64
    Description string
    DislikeCount uint64
    Duration string
    FavoriteCount uint64
    LikeCount uint64
    PublishedAt string
    Tags string
    Title string
    ViewCount uint64
}
