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
	go client.InitWriteClient(5, 8085, "ChunkingTest4.txt", "shared_chunk.txt")
	go client.InitWriteClient(6, 8086, "ChunkingTest5.txt", "shared_chunk.txt")
	go client.InitWriteClient(7, 8087, "ChunkingTest6.txt", "shared_chunk.txt")
	go client.InitWriteClient(8, 8088, "ChunkingTest7.txt", "shared_chunk.txt")
	go client.InitWriteClient(9, 8089, "ChunkingTest8.txt", "shared_chunk.txt")
	go client.InitWriteClient(10, 8070, "ChunkingTest9.txt", "shared_chunk.txt")

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
	go client.InitWriteClient(7, 8087, "ChunkingTest.txt", "shared_chunk.txt")
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

func main() {
	// Test for the Append Fucntion.
	// One write by a single client.
	// singleWrite()

	// Multiple concurrent writes by a single client.
	// chunkingWrites()

	// Concurrent Writes by multiple clients.
	concurrentWrites()

	// Read operation by a client.
	// readChunk()
}
