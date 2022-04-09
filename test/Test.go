package main

import (
	helper "oks/Helper"
	// "fmt"
	// "os/exec"
)

// client "oks/Client"

func main() {
	// go client.InitClient(7, 8087, "test2.txt", "shared_chunk.txt")
	// go client.InitClient(8, 8088, "test1.txt", "shared_chunk.txt")
	// for {}
	helper.RunClient(7, 8087, "test2.txt", "shared_chunk.txt")
	helper.RunChunkServer(1, 8081)
	helper.RunChunkServer(2, 8082)
	helper.RunChunkServer(3, 8083)
	helper.RunChunkServer(4, 8084)
	helper.RunChunkServer(5, 8085)
	helper.RunMaster()
}
