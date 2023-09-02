package main

import "yourtube.51390.cloud/yourtube"

func main() {
    db, err := yourtube.InitDb()
    yourtube.HandleError(err, "Failed initializing db")
    yourtube.Migrate(db)
    yourtube.Sync(db)
}
