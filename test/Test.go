package main

import (
	client "oks/Client"
	time "time"
)

func singleWrite() {
	go client.InitClient(7, 8087, "test2.txt", "shared_chunk.txt")
	for {}
}

func concurrentWrites() {
	go client.InitClient(7, 8087, "test2.txt", "shared_chunk.txt")
	go client.InitClient(8, 8088, "test1.txt", "shared_chunk.txt")
	for {}
}

func consecutiveWrites() {
	go client.InitClient(7, 8087, "test2.txt", "shared_chunk.txt")
	time.Sleep(time.Second * 5)
	go client.InitClient(8, 8088, "test1.txt", "shared_chunk.txt")
	for {}
}

func main() {
	singleWrite()
}
