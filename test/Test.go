package main

import (
	client "oks/Client"
	time "time"
)

func singleWrite() {
	go client.InitWriteClient(7, 8087, "test1.txt", "shared_chunk.txt")
	for {
	}
}

func concurrentWrites() {
	go client.InitWriteClient(1, 8081, "ChunkingTest0.txt", "shared_chunk.txt")
	go client.InitWriteClient(2, 8082, "ChunkingTest1.txt", "shared_chunk.txt")
	go client.InitWriteClient(3, 8083, "ChunkingTest2.txt", "shared_chunk.txt")
	go client.InitWriteClient(4, 8084, "ChunkingTest3.txt", "shared_chunk.txt")
	//go client.InitWriteClient(5, 8089, "test0.txt", "shared_chunk.txt")

	for {
	}
}

func consecutiveWrites() {
	go client.InitWriteClient(7, 8087, "test2.txt", "shared_chunk.txt")
	time.Sleep(time.Second * 5)
	go client.InitWriteClient(8, 8088, "test1.txt", "shared_chunk.txt")
	for {
	}
}

func chunkingWrites() {
	go client.InitWriteClient(7, 8087, "test.txt", "shared_chunk.txt")
	for {
	}
}

func readChunk() {
	go client.InitReadClient(1, 8081, "shared_chunk.txt_c0")
	go client.InitReadClient(2, 8082, "shared_chunk.txt_c0")
	go client.InitReadClient(3, 8083, "shared_chunk.txt_c0")
	go client.InitReadClient(4, 8084, "shared_chunk.txt_c0")
	go client.InitReadClient(5, 8085, "shared_chunk.txt_c0")
	go client.InitReadClient(6, 8086, "shared_chunk.txt_c0")

	for {
	}
}

//
func main() {
	//singleWrite()
	//chunkingWrites()

	concurrentWrites()
	// readChunk()
}

// uint64=13875145260 088531144
// uint64=13875145260 785567868 end
