package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	//client "oks/Client"
	"math/rand"
	chunkServer "oks/ChunkServer"
	helper "oks/Helper"
	structs "oks/Structs"
	"reflect"
	"strconv"
)

var metaData MetaData
var port_map Port

type heartState int

const (
	Start heartState = iota
	Pending
	Alive
	Dead
)

type MetaData struct {
	// key: file id int, value: chunk array
	// eg file 1 = [f1_c0, file1_chunk2, file1_chunk3]
	fileIdToChunkId map[string][]string

	// map each file chunk to a chunk server (port number)
	chunkIdToChunkserver map[string][]int

	// map each chunkserver to the amount of storage it has
	// max chunk is 10 KB
	// max append data is 2.5KB
	// {f1 : {c0 : 0KB, c1 : 2KB} }
	chunkIdToOffset map[string]int64

	heartBeatAck sync.Map
	// {8080:{f1_c0}}
	chunkServerToChunkId map[int][]string
}

type Port struct {
	portToInt map[string]int
}

func listen(nodePid int, portNo int) {
	router := gin.Default()
	// router.Use(gin.CustomRecovery(func (c *gin.Context, recovered interface{}){
	// 	if err, ok := recovered.(string); ok {
	// 		c.String(http.StatusInternalServerError, fmt.Sprintf("error: %s", err))
	// 	}
	// }))
	router.POST("/message", postMessageHandler)

	fmt.Printf("Node %d listening on port %d \n", nodePid, portNo)
	router.Run("localhost:" + strconv.Itoa(portNo))
}

func postMessageHandler(context *gin.Context) {
	var message structs.Message

	// Call BindJSON to bind the received JSON to message.
	if err := context.BindJSON(&message); err != nil {
		fmt.Println("Invalid message object received.")
		return
	}

	if message.MessageType != helper.ACK_HEARTBEAT {
		context.IndentedJSON(http.StatusOK, message.MessageType+" message from Node "+strconv.Itoa(message.Ports[message.Pointer])+" was received by Master")
	} else {
		context.IndentedJSON(http.StatusOK, message.MessageType+" message from Node "+strconv.Itoa(message.Ports[message.Pointer+1])+" was received by Master")
	}
	switch message.MessageType {
	case helper.DATA_APPEND:
		go appendMessageHandler(message)
	case helper.ACK_CHUNK_CREATE:
		go ackChunkCreate(message)
	case helper.ACK_HEARTBEAT:
		go receiveHeartbeatACK(message)
	}
}

func appendMessageHandler(message structs.Message) {
	fmt.Println("Before Sending")
	if _, ok := metaData.fileIdToChunkId[message.Filename]; !ok {
		// if file does not exist in metaData, create a new entry
		fmt.Println("Master received new file from Client")
		newFileAppend(message)
	} else {
		// if file is not new
		fmt.Println("Master received old file from Client")
		fileNotNew(message)

	}
}

func fileNotNew(message structs.Message) {
	//create message to send to client
	chunkId := metaData.fileIdToChunkId[message.Filename][len(metaData.fileIdToChunkId[message.Filename])-1]
	messagePorts := message.Ports // [C, M]
	messagePorts = append([]int{messagePorts[0]}, metaData.chunkIdToChunkserver[chunkId]...)
	// [C, P, S1, S2]

	message1 := structs.Message{
		MessageType: helper.DATA_APPEND,
		// master, primary, secondary_1, secondary_2
		Ports:          messagePorts, // [C, P, S1, S2]
		Pointer:        0,
		SourceFilename: message.SourceFilename,
		Filename:       message.Filename,
		ChunkId:        chunkId,
		Payload:        message.Payload,
		PayloadSize:    message.PayloadSize,
		ChunkOffset:    metaData.chunkIdToOffset[chunkId],
		RecordIndex:    message.RecordIndex,
	}
	fmt.Println("Master sending request to primary chunkserver")
	helper.SendMessage(message1)

	// increment offset
	metaData.chunkIdToOffset[chunkId] += message1.PayloadSize
}

// Client wants to append to a new file
func newFileAppend(message structs.Message) {
	// create new entry in MetaData
	chunkId := message.Filename + "_c0"
	metaData.fileIdToChunkId[message.Filename] = []string{chunkId}

	// ask 3 chunkserver to create chunks
	fmt.Println("Master choosing 3 chunkservers")
	new_chunkServer := choose_3_random_chunkServers()
	messagePorts := message.Ports // [C, M]
	messagePorts = append(messagePorts, new_chunkServer...)

	message1 := structs.Message{
		MessageType: helper.CREATE_NEW_CHUNK,
		// master, primary, secondary_1, secondary_2
		Ports:          messagePorts, // [C, M, P, S1, S2]
		Pointer:        2,
		SourceFilename: message.SourceFilename,
		Filename:       message.Filename,
		ChunkId:        chunkId,
		Payload:        message.Payload,
		PayloadSize:    message.PayloadSize,
		ChunkOffset:    0,
		RecordIndex:    message.RecordIndex,
	}
	fmt.Println("Master sending request to primary chunkserver")
	helper.SendMessage(message1)

}

// after receiving ack from primary, approve request for client
func ackChunkCreate(message structs.Message) {
	messagePorts := message.Ports // [C, M, P, S1, S2]
	fmt.Println("[C, M, P, S1, S2]", messagePorts)
	messagePorts = append([]int{messagePorts[0]}, messagePorts[2:]...) //[C, P, S1, S2]
	fmt.Println("CHUNK CREATED - PRIOR UPDATING PORTS [C, P, S1, S2]", messagePorts)
	chunkServers := messagePorts[1:]
	fmt.Println(chunkServers)

	message1 := structs.Message{
		MessageType:    helper.DATA_APPEND,
		Ports:          messagePorts, // [C, P, S1, S2]
		Pointer:        0,
		SourceFilename: message.SourceFilename,
		Filename:       message.Filename,
		ChunkId:        message.ChunkId,
		Payload:        message.Payload,
		PayloadSize:    message.PayloadSize,
		ChunkOffset:    0,
		RecordIndex:    message.RecordIndex,
	}
	fmt.Println("Master approving append request to client")
	helper.SendMessage(message1)

	//record in metaData
	for _, v := range chunkServers {
		metaData.chunkServerToChunkId[v] = append(metaData.chunkServerToChunkId[v], message.ChunkId)
	}
	metaData.chunkIdToChunkserver[message.ChunkId] = chunkServers
	fmt.Println(metaData.chunkIdToChunkserver[message.ChunkId])

	// increment offset
	new_offset := metaData.chunkIdToOffset[message.ChunkId] + message1.PayloadSize
	metaData.chunkIdToOffset[message.ChunkId] = new_offset
}
func (m heartState) String() string {
	return [...]string{"START", "PENDING", "ALIVE", "DEAD"}[m]
}

// Separate go-routine to send heartbeat messages in intervals
func sendHeartbeat() {

	for {
		metaData.printACKMap()
		for i := 8081; i <= 8085; i++ {
			currentHeartState, _ := metaData.heartBeatAck.Load(i)
			switch currentHeartState {
			case Start: // For the first heartbeat
				metaData.heartBeatAck.Store(i, Pending)
				heartbeatMsg := structs.Message{
					MessageType:    helper.HEARTBEAT,
					Ports:          []int{helper.MASTER_SERVER_PORT, i}, // [C, P, S1, S2]
					Pointer:        1,
					SourceFilename: "",
					Filename:       "",
					ChunkId:        "",
					Payload:        "",
					PayloadSize:    0,
					ChunkOffset:    0,
					RecordIndex:    1,
				}
				go helper.SendMessage(heartbeatMsg)
			case Pending: // No reply from the previous heartbeat, consider the chunk server dead
				metaData.heartBeatAck.Store(i, Dead)
			case Alive: // Successfully received ack from previous heartbeat, send next heartbeat
				metaData.heartBeatAck.Store(i, Pending)
				heartbeatMsg := structs.Message{
					MessageType:    helper.HEARTBEAT,
					Ports:          []int{helper.MASTER_SERVER_PORT, i}, // [C, P, S1, S2]
					Pointer:        1,
					SourceFilename: "",
					Filename:       "",
					ChunkId:        "",
					Payload:        "",
					PayloadSize:    0,
					ChunkOffset:    0,
					RecordIndex:    1,
				}
				go helper.SendMessage(heartbeatMsg)
			case Dead: // Chunk server is dead, do nothing
			}

		}

		time.Sleep(time.Second * 5)
	}
}

// Updates heartbeat map to true if receive heartbeat ack from chunk server
func receiveHeartbeatACK(message structs.Message) {
	//fmt.Println(message.Ports[1], message)
	metaData.heartBeatAck.Store(message.Ports[1], Alive)
}

// @ts-ignore
func (m *MetaData) printACKMap() {
	//fmt.Println(m.heartBeatAck)
	// Ignore any error from here, works just fine
	m.heartBeatAck.Range(func(k, v interface{}) bool {
		fmt.Println(k, v)
		return true
	})
}

func choose_3_random_chunkServers() []int {

	chunkServerArray := map[int]bool{
		8081: false,
		8082: false,
		8083: false,
		8084: false,
		8085: false,
	}

	res := []int{}

	for len(res) < 3 {
		//random key stores the key from the chunkS
		random_key := MapRandomKeyGet(chunkServerArray).(int)
		// checking if this key boolean is false or true, if false append this key to the res and set the key value true instead
		if chunkServerArray[random_key] == false {
			chunkServerArray[random_key] = true
			res = append(res, random_key)
			fmt.Println(res)

		} else {
			//if the chunkS[random_key]==true, it means that the random key has been added into the res array
			continue
		}
	}
	return res

}

//this will select random keys in the map
func MapRandomKeyGet(mapI interface{}) interface{} {
	keys := reflect.ValueOf(mapI).MapKeys()

	return keys[rand.Intn(len(keys))].Interface()
}

// Create ACKMap on start-up
func (m *MetaData) initialiseACKMap() {
	for i := 8081; i <= 8085; i++ {
		m.heartBeatAck.Store(i, Start)
	}
}

func main() {
	metaData.fileIdToChunkId = make(map[string][]string)
	metaData.chunkIdToChunkserver = make(map[string][]int)
	metaData.chunkServerToChunkId = make(map[int][]string)
	metaData.chunkIdToOffset = make(map[string]int64)
	metaData.initialiseACKMap()
	metaData.printACKMap()

	port_map.portToInt = map[string]int{"0": 8080, "1": 8081, "2": 8082, "3": 8083, "4": 8084, "5": 8085}

	// create dummy data
	// metaData.fileIdToChunkId["test.txt"] = []string{"test_c0"}
	// metaData.chunkIdToChunkserver["test_c0"] = []int{8081, 8082, 8083}
	// offset := int64(0)
	// metaData.chunkIdToOffset["test_c0"] = offset

	go chunkServer.ChunkServer(1, 8081)
	go chunkServer.ChunkServer(2, 8082)
	go chunkServer.ChunkServer(3, 8083)
	go chunkServer.ChunkServer(4, 8084)
	go chunkServer.ChunkServer(5, 8085)
	go sendHeartbeat()
	listen(0, 8080)

}

// TODO

// Priority: Best case [ DONE ]
// ASSUME ALL REPLY for now
// 1. Send Heartbeat to ChunkServers
// 2. Listen for Heartbeat
// 3. Update HeartbeatAck accordingly

// Next case: Dead chunk server
// Assume a node fails, master does not receive ACK
// 1. Implement states
// 2. Update HeartbeatAck

// 3. Check metadata of chunk server that failed
// 		3a. Get ALL chunk ID from chunkServerToChunkId
// 		3b. Get available chunk servers for each chunk ID, check chunkIdToChunkserver
// 		3c. Get chunk servers that do not have chunk, check chunkIdToChunkserver
// 		3d. Select which chunk server you want to replicate to
// 		3e.
