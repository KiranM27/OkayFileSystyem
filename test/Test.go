package main

import client "oks/Client"

// import "time"

func main() {
	client.InitClient(7, 8087, "test2.txt", "shared_chunk.txt")
}
