package main

import (
	"bdecoder/bdecoder"
	"fmt"
	"os"
)

func main() {
	bencodedString, err := os.ReadFile("torrents/example.torrent")
	if err != nil {
		fmt.Println("Error reading file:", err)
		os.Exit(1)
	}
	bencodedTorrent, err := bdecoder.Decode(string(bencodedString))
	if err != nil {
		fmt.Println("Error decoding object:", err)
		os.Exit(1)
	}
	fmt.Println(bencodedTorrent)
}
