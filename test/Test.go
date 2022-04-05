package main

import (
	client "oks/Client"
)

func main() {
	go client.InitClient(7, 8087, "test2.txt", "shared_chunk.txt")
	go client.InitClient(8, 8088, "test1.txt", "shared_chunk.txt")
	for {}
}
