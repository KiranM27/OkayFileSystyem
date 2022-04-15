package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"errors"
	"math/rand"
	chunkServer "oks/ChunkServer"
	helper "oks/Helper"
	structs "oks/Structs"
	"reflect"
	"strconv"

	"github.com/gin-gonic/gin"
)

var metaData MetaData
var lock sync.Mutex

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
	replicationMap       map[int][]structs.RepMsg
	successfulWrites     map[string][]structs.SuccessfulWrite
}

// Placeholder - TODO: Add proper replication message struct to message.go

func listen(nodePid int, portNo int) {
	router := gin.Default()
	router.POST("/message", postMessageHandler)
	router.POST("/replicate", repMessageHandler)
	router.POST("/read", readMessagehandler)
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
	case helper.ACK_COMMIT:
		go ackCommitHandler(message)
	case helper.REVIVE: // same thing as get heartbeat
		go receiveHeartbeatACK(message)
	}
}

func repMessageHandler(context *gin.Context) {
	var repMsg structs.RepMsg

	// Call BindJSON to bind the received JSON to message.
	if err := context.BindJSON(&repMsg); err != nil {
		fmt.Println("Invalid message object received.", err)
		return
	}
	context.IndentedJSON(http.StatusOK, repMsg.MessageType+" Received")
	fmt.Println("---------- Received Replication Message: ", repMsg.MessageType, " ----------")

	switch repMsg.MessageType {
	case helper.ACK_REPLICATION:
		receiveReplicationACK(repMsg)
	}
}

func readMessagehandler(context *gin.Context) {
	var readMsg structs.ReadMsg

	// Call BindJSON to bind the received JSON to message.
	if err := context.BindJSON(&readMsg); err != nil {
		fmt.Println("Invalid message object received.", err)
		return
	}
	responseMsg := readHandler(readMsg)
	context.IndentedJSON(http.StatusOK, responseMsg)
}

func ackCommitHandler(message structs.Message) {
	lock.Lock()
	SWObj := structs.GenerateSW(message.ChunkOffset, message.ChunkOffset+message.PayloadSize)
	metaData.successfulWrites[message.ChunkId] = append(metaData.successfulWrites[message.ChunkId], SWObj)
	lock.Unlock()
}

func appendMessageHandler(message structs.Message) {
	fmt.Println("Before Sending")
	if _, ok := metaData.fileIdToChunkId[message.Filename]; !ok {
		// if file does not exist in metaData, create a new entry
		// fmt.Println("Master received new file from Client")
		newFileAppend(message)
	} else {
		// if file is not new
		// fmt.Println("Master received old file from Client")
		fmt.Println("CHECK THIS ",  metaData.fileIdToChunkId[message.Filename])
		fileNotNew(message)

	}
}

func readHandler(readMsg structs.ReadMsg) structs.ReadMsg {
	chunkId := readMsg.ChunkId
	responseMsg := structs.GenerateReadMsg(helper.ACK_READ_REQ, chunkId, metaData.chunkIdToChunkserver[chunkId], metaData.successfulWrites[chunkId], "")
	return responseMsg
}

func fileNotNew(message structs.Message) {
	fmt.Println("FUCK ", len(metaData.fileIdToChunkId[message.Filename]))

	chunkId := metaData.fileIdToChunkId[message.Filename][len(metaData.fileIdToChunkId[message.Filename])-1]
	fmt.Println("NOT HEER")
	if helper.CHUNK_SIZE - metaData.chunkIdToOffset[chunkId] < message.PayloadSize + 1 {
		fmt.Println("HERE")
		newChunkNo := len(metaData.fileIdToChunkId[message.Filename])
		newChunkId := message.Filename + "_c" + strconv.Itoa(newChunkNo)
		createNewFile(newChunkId, message)
	}

	//create message to send to client
	messagePorts := message.Ports // [C, M]
	messagePorts = append([]int{messagePorts[0]}, metaData.chunkIdToChunkserver[chunkId]...)
	// [C, P, S1, S2]
	if len(messagePorts) != 4 { // doesn't reply to client as file creation is ongoing.
		return
	}

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
	// fmt.Println("Master sending request to primary chunkserver")
	helper.SendMessage(message1)

	// increment offset
	lock.Lock()
	metaData.chunkIdToOffset[chunkId] += message1.PayloadSize
	lock.Unlock()
}

// Client wants to append to a new file
func newFileAppend(message structs.Message) {
	// create new entry in MetaData
	chunkId := message.Filename + "_c0"
	createNewFile(chunkId, message)
}

func createNewFile(chunkId string, message structs.Message) {
	lock.Lock()
	tempChunkList := []string{}
	if val, ok := metaData.fileIdToChunkId[message.Filename]; ok {
		tempChunkList = append(tempChunkList, val...)
	}
	fmt.Println(tempChunkList)
	metaData.fileIdToChunkId[message.Filename] = append(tempChunkList, chunkId)
	lock.Unlock()
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
	// fmt.Println("[C, M, P, S1, S2]", messagePorts)
	messagePorts = append([]int{messagePorts[0]}, messagePorts[2:]...) //[C, P, S1, S2]
	// fmt.Println("CHUNK CREATED - PRIOR UPDATING PORTS [C, P, S1, S2]", messagePorts)
	chunkServers := messagePorts[1:]
	// fmt.Println(chunkServers)

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
	lock.Lock()
	for _, v := range chunkServers {
		metaData.chunkServerToChunkId[v] = append(metaData.chunkServerToChunkId[v], message.ChunkId)
	}
	metaData.chunkIdToChunkserver[message.ChunkId] = chunkServers
	fmt.Println(metaData.chunkIdToChunkserver[message.ChunkId])

	// increment offset
	new_offset := metaData.chunkIdToOffset[message.ChunkId] + message1.PayloadSize
	metaData.chunkIdToOffset[message.ChunkId] = new_offset
	lock.Unlock()
}

func (m heartState) String() string {
	return [...]string{"START", "PENDING", "ALIVE", "DEAD"}[m]
}

// Separate go-routine to send heartbeat messages in intervals
func sendHeartbeat() {

	for {
		// metaData.printACKMap()
		for i := 8081; i <= 8085; i++ {
			currentHeartState, _ := metaData.heartBeatAck.Load(i)
			switch currentHeartState {
			case Start: // For the first heartbeat
				metaData.heartBeatAck.Store(i, Pending)
				heartbeatMsg := structs.Message{
					MessageType: helper.HEARTBEAT,
					Ports:       []int{helper.MASTER_SERVER_PORT, i}, // [C, P, S1, S2]
					Pointer:     1,
				}
				go helper.SendMessage(heartbeatMsg)
			case Pending: // No reply from the previous heartbeat, consider the chunk server dead
				// TODO: SEND HERE
				go startReplicate(i) // give the port number to function
				metaData.heartBeatAck.Store(i, Dead)
			case Alive: // Successfully received ack from previous heartbeat, send next heartbeat
				metaData.heartBeatAck.Store(i, Pending)
				heartbeatMsg := structs.Message{
					MessageType: helper.HEARTBEAT,
					Ports:       []int{helper.MASTER_SERVER_PORT, i}, // [C, P, S1, S2]
					Pointer:     1,
				}
				go helper.SendMessage(heartbeatMsg)
			case Dead: // Chunk server is dead, replicate server
				fmt.Println(i, " is dead!")
			}

		}

		time.Sleep(time.Second * 5)
	}
}

func startReplicate(chunkServerId int) error {
	lock.Lock()
	temporaryReplicationMap := make(map[int][]structs.RepMsg)

	// 3a - Get the list of chunks in the failed chunk server
	failedChunkIds := metaData.chunkServerToChunkId[chunkServerId]

	// 3e - Remove chunkServerId from the metadata first
	for _, failedChunkId := range failedChunkIds { // for each chunk in the failed server
		for idx, serverId := range metaData.chunkIdToChunkserver[failedChunkId] { // iterate through list of chunk servers associated with chunk id
			if serverId == chunkServerId { // if this is the chunk server id, remove it
				if len(metaData.chunkIdToChunkserver[failedChunkId]) == 1 { // edge case, only element in list
					return errors.New("Write Failed for Replication of chunk " + failedChunkId)
				} else {
					fmt.Println("REMOVING THE FAILED CHUNK NOW")
					metaData.chunkIdToChunkserver[failedChunkId] = append(metaData.chunkIdToChunkserver[failedChunkId][:idx], metaData.chunkIdToChunkserver[failedChunkId][idx+1:]...)
					fmt.Println(metaData.chunkIdToChunkserver[failedChunkId])
				}
				break
			}
		}
	}

	deadServerChunks := []structs.RepMsg{}

	// if replication failed, get the chunk ids that it was supposed to replicate
	if val, ok := metaData.replicationMap[chunkServerId]; ok {
		deadServerChunks = append(deadServerChunks, val...)
		for _, failedRepMsgStruct := range val {
			failedChunkIds = append(failedChunkIds, failedRepMsgStruct.ChunkId)
		}
	}

	// Get metadata needed 3b
	for _, failedChunkId := range failedChunkIds { // iterate through chunk ids from failed chunk server
		sourceServers := metaData.chunkIdToChunkserver[failedChunkId] // get list of remaining chunk servers associated to chunk ID

		// 3c - Get available servers we can replicate to
		// Comment: Do we have to know every available one? Once we find one lets just pick it and end.
		var replicateServer int
		metaData.heartBeatAck.Range(func(replicateCandidate, replicateCandidateStatus interface{}) bool {
			tempRepMapFlag := false
			if !Contains(sourceServers, replicateCandidate.(int)) && replicateCandidateStatus.(heartState) != Dead { // check if the chunk server id is in sourceServers, if yes means it already has chunk ID
				for _, replicateCandidateRepMsgStruct := range deadServerChunks {
					if replicateCandidateRepMsgStruct.TargetCS == replicateCandidate {
						tempRepMapFlag = true
					}
				}
				if !tempRepMapFlag {
					replicateServer = replicateCandidate.(int)
					return false
				}
			}
			return true
		})

		// create new replication message
		newReplication := structs.RepMsg{
			Index:       len(metaData.replicationMap[replicateServer]),
			MessageType: helper.REPLICATE,
			ChunkId:     failedChunkId, // Comment - use this ineted of index
			Sources:     sourceServers,
			TargetCS:    replicateServer,
		}

		// 3f - add replication request to replication map
		metaData.replicationMap[replicateServer] = append(metaData.replicationMap[replicateServer], newReplication)
		temporaryReplicationMap[replicateServer] = append(temporaryReplicationMap[replicateServer], newReplication)
	}
	lock.Unlock()
	// 3g - Send to target [NOT DONE]
	// TODO - New SendMessage for Replication Message
	for _, replicationMsgs := range temporaryReplicationMap {
		for _, repMessage := range replicationMsgs {
			fmt.Println(repMessage.ChunkId, repMessage.TargetCS)
			go helper.SendRepMsg(repMessage, repMessage.TargetCS) // need new helper function for rep msg
		}
	}
	return nil
}

// 3h
func receiveReplicationACK(RepMsg structs.RepMsg) {
	lock.Lock()
	defer lock.Unlock()
	for idx, repRequest := range metaData.replicationMap[RepMsg.TargetCS] {
		if repRequest.Index == RepMsg.Index {
			if len(metaData.replicationMap[RepMsg.TargetCS]) == 1 { // edge case, only request in replication map
				metaData.replicationMap[RepMsg.TargetCS] = metaData.replicationMap[RepMsg.TargetCS][:1]
			} else {
				metaData.replicationMap[RepMsg.TargetCS] = append(metaData.replicationMap[RepMsg.TargetCS][:idx], metaData.replicationMap[RepMsg.TargetCS][idx+1:]...)
			}
			break
		}
	}

	// 3i - update metadata
	chunkServerId := RepMsg.TargetCS
	chunkId := RepMsg.ChunkId

	// Update {chunkId: {chunkserver1, chunkserver2, ...}}
	metaData.chunkIdToChunkserver[chunkId] = append(metaData.chunkIdToChunkserver[chunkId], chunkServerId)

	// Update {chunkserver1: {chunkid1, chunkid2, ...}}
	metaData.chunkServerToChunkId[chunkServerId] = append(metaData.chunkServerToChunkId[chunkServerId], chunkId)
}

// Updates heartbeat map to true if receive heartbeat ack from chunk server
func receiveHeartbeatACK(message structs.Message) {
	metaData.heartBeatAck.Store(message.Ports[1], Alive)
}

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
	metaData.replicationMap = make(map[int][]structs.RepMsg)
	metaData.successfulWrites = make(map[string][]structs.SuccessfulWrite)
	metaData.initialiseACKMap()

	go chunkServer.ChunkServer(1, 8081)
	go chunkServer.ChunkServer(2, 8082)
	go chunkServer.ChunkServer(3, 8083)
	go chunkServer.ChunkServer(4, 8084)
	go chunkServer.ChunkServer(5, 8085)
	go sendHeartbeat()
	listen(0, 8080)

}

/*
// Priority: Best case [ DONE ]
// ASSUME ALL REPLY for now
// 1. Send Heartbeat to ChunkServers
// 2. Listen for Heartbeat
// 3. Update HeartbeatAck accordingly

// Next case: Dead chunk server
// Assume a node fails, master does not receive ACK
// 1. Implement states
// 2. Update HeartbeatAck

 3. Check metadata of chunk server that failed
 		3a. Get ALL chunk ID from chunkServerToChunkId
 			3b. Get available chunk servers for each chunk ID, check chunkIdToChunkserver (c1: [chunkserver1, chunkserver2])
			3c. Get chunk servers that do not have chunk, check chunkIdToChunkserver (c1: [chunkserver3, chunkserver4])
			3d. Select which chunk server you want to replicate to. (ChunkId : c1, Target: chunkserver3, Sources: chunkserver1, chunkserver2)
 			3e. Remove chunkId from chunkIdToChunkserver and chunkIdToChunkserver
		3f. Add struct to rep map (Chunkserver3 : [(ChunkId : c1, Sources: chunkserver1, chunkserver2), ...]
		3g. Send to target Chunkserver
		3h. After receiving ACK REPLICATION from chunkserver 3, remove entry from rep map
		3i. Update Metadata
*/
