package main

import (
	"fmt"
	"github.com/jkittell/mediastreamdownloader/downloader"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		log.Println("Please provide the url of the stream to download")
		os.Exit(1)
	}
	url := os.Args[1]
	res := downloader.Run(url)
	for i := 0; i < res.Length(); i++ {
		str := res.Lookup(i)
		data, err := str.JSON()
		if err != nil {
			log.Println(err)
		} else {
			fmt.Println(string(data))
		}
	}
}
