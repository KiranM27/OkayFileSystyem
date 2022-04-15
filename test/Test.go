package main

import (
	client "oks/Client"
	time "time"
)

func singleWrite() {
	go client.InitWriteClient(7, 8087, "test.txt", "shared_chunk.txt")
	for {}
}

func concurrentWrites() {
	go client.InitWriteClient(7, 8087, "test2.txt", "shared_chunk.txt")
	go client.InitWriteClient(8, 8088, "test1.txt", "shared_chunk.txt")
	for {}
}

func consecutiveWrites() {
	go client.InitWriteClient(7, 8087, "test2.txt", "shared_chunk.txt")
	time.Sleep(time.Second * 5)
	go client.InitWriteClient(8, 8088, "test1.txt", "shared_chunk.txt")
	for {}
}

func readChunk() {
	go client.InitReadClient(7, 8087, "shared_chunk.txt_c0")
	for {}
}

func main() {
	singleWrite()
	// readChunk()
}
