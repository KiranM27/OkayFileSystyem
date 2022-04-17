package main

import (
	client "oks/Client"
	helper "oks/Helper"
	"fmt"
	"time"
)

func InfiniteLoop() {
	for {}
}

func SingleWriteTest() {
	go client.InitWriteClient(7, 8087, "ChunkingTest0.txt", "shared_chunk.txt")
	InfiniteLoop()
}

func ConcurrentWritesTest(noOfClients int) {
	CLIENT_START_PORT := helper.CLIENT_START_PORT
	for i := 0; i < noOfClients; i++ {
		go client.InitWriteClient(i, CLIENT_START_PORT + i, fmt.Sprintf("ChunkingTest%d.txt", i), "shared_chunk.txt")
	}
	InfiniteLoop()
}

func ConsecutiveWritesTest() {
	go client.InitWriteClient(7, 8087, "ChunkingTest0.txt", "shared_chunk.txt")
	time.Sleep(time.Second * 2)
	go client.InitWriteClient(8, 8088, "ChunkingTest1.txt", "shared_chunk.txt")
	InfiniteLoop()
}

func ChunkingWritesTest() {
	go client.InitWriteClient(7, 8087, "ChunkingTest.txt", "shared_chunk.txt")
	InfiniteLoop()
}

func ReadChunkTest(noOfClients int) {
	CLIENT_START_PORT := helper.CLIENT_START_PORT
	for i := 0; i < noOfClients; i++ {
		go client.InitReadClient(i, CLIENT_START_PORT + i, "shared_chunk_c0")
	}
	InfiniteLoop()
}

func main() {
	// Test for the Append Fucntion.
	// One write by a single client.
	// SingleWriteTest()

	// Multiple concurrent writes by a single client.
	// ChunkingWritesTest()

	// Concurrent Writes by multiple clients.
	ConcurrentWritesTest(10)

	// Read operation by a client.
	// ReadChunkTest()
}
