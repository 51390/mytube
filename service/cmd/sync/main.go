package main

import "mytube.51390.cloud/mytube"

func main() {
	mytube.LoadEnv()
	mytube.Sync("", "", "")
}
