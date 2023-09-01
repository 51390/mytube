package main

import "yourtube.51390.cloud/yourtube"

func main() {
    _, err := yourtube.InitDb()
    yourtube.HandleError(err, "Failed initializing db")
    yourtube.Sync()
}
