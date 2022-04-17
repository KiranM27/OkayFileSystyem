package main

import (
	"fmt"
	client "oks/Client"
	helper "oks/Helper"
	"strconv"
	"strings"
	"time"
	"math"
)

func InfiniteLoop() {
	for {}
}

func CreateLogFiles() {
	helper.CreateFile(helper.START_TIMES_LOG_FILE)
	helper.CreateFile(helper.END_TIMES_LOG_FILE)
}

func AtoIArray(_array []string) []int {
	var output []int
	for _, val := range _array {
		val = strings.TrimSpace(val)
		intTime, err := strconv.Atoi(val)
		if err == nil {
			output = append(output, intTime)
		}
	}
	return output
}

func MinMax(array []int) (int, int) {
	var max int = array[0]
	var min int = array[0]
	for _, value := range array {
		if max < value {
			max = value
		}
		if min > value {
			min = value
		}
	}
	return min, max
}

func DetermineOperationTime() {
	startTimesContent := helper.ReadFile(helper.START_TIMES_LOG_FILE)
	endTimesContent := helper.ReadFile(helper.END_TIMES_LOG_FILE)
	startTimes := AtoIArray(strings.Split(startTimesContent, "\n"))
	endTimes := AtoIArray(strings.Split(endTimesContent, "\n"))
	startTime, _ := MinMax(startTimes)
	_, endTime := MinMax(endTimes)
	duration := float64((endTime - startTime))
	duration = duration / math.Pow(10, 9)
	fmt.Println("The duration of the entire operation was ", duration, " s")
}

func SingleWriteTest() {
	go client.InitWriteClient(7, 8087, "ChunkingTest0.txt", "shared_chunk.txt")
	InfiniteLoop()
}

func ConcurrentWritesTest(noOfClients int) {
	CLIENT_START_PORT := helper.CLIENT_START_PORT
	for i := 0; i < noOfClients; i++ {
		go client.InitWriteClient(i, CLIENT_START_PORT+i, fmt.Sprintf("ChunkingTest%d.txt", i), "shared_chunk.txt")
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
		go client.InitReadClient(i, CLIENT_START_PORT+i, "shared_chunk_c0")
	}
	InfiniteLoop()
}

func main() {
	// CreateLogFiles()

	// Test for the Append Fucntion.
	// One write by a single client.
	// SingleWriteTest()

	// Multiple concurrent writes by a single client.
	// ChunkingWritesTest()

	// Concurrent Writes by multiple clients.
	// ConcurrentWritesTest(10)

	// Read operation by a client.
	// ReadChunkTest(4)

	// Determine the duration of the operation.
	DetermineOperationTime()
}
