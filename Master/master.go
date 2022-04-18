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
var liveChunkServers = []int{}
var fullCapacityServers = []int{}
var latestChunkServer = 8090

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


func listen(nodePid int, portNo int) {
	router := gin.New()
    router.Use(
        gin.LoggerWithWriter(gin.DefaultWriter, "/message", "/replicate", "/read"),
        gin.Recovery(),
    )
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
	fmt.Println("Master has received a append request from client - ", message.Ports[0], " - for file - ", message.SourceFilename)
	if _, ok := metaData.fileIdToChunkId[message.Filename]; !ok {
		// File does not exist in metaData, create a new entry
		fmt.Println("File doesn't exist in OFS. Creating the file - ", message.SourceFilename, " - in OFS now.")
		newFileAppend(message)
	} else {
		// File exists in the OFS, try to append to it
		fmt.Println("File exists in OFS. Will try to append to file - ", message.SourceFilename, " - in OFS now.")
		fileNotNew(message)
	}
}

func readHandler(readMsg structs.ReadMsg) structs.ReadMsg {
	chunkId := readMsg.ChunkId
	responseMsg := structs.GenerateReadMsg(helper.ACK_READ_REQ, chunkId, metaData.chunkIdToChunkserver[chunkId], metaData.successfulWrites[chunkId], "")
	return responseMsg
}

func fileNotNew(message structs.Message) {
	chunkId := metaData.fileIdToChunkId[message.Filename][len(metaData.fileIdToChunkId[message.Filename])-1]
	fileName := helper.RemoveExtension(message.Filename)
	if helper.CHUNK_SIZE-metaData.chunkIdToOffset[chunkId] <= message.PayloadSize {
		newChunkNo := len(metaData.fileIdToChunkId[message.Filename])
		newChunkId := fileName + "_c" + strconv.Itoa(newChunkNo)
		createNewFile(newChunkId, message)
		return
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
		// client, primary, secondary_1, secondary_2
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
	helper.SendMessage(message1)

	// increment offset
	lock.Lock()
	metaData.chunkIdToOffset[chunkId] += message1.PayloadSize
	lock.Unlock()
}

// Client wants to append to a new file
func newFileAppend(message structs.Message) {
	fileName := helper.RemoveExtension(message.Filename)
	chunkId := fileName + "_c0"
	createNewFile(chunkId, message)
}

func createNewFile(chunkId string, message structs.Message) {
	lock.Lock()
	tempChunkList := []string{}
	if val, ok := metaData.fileIdToChunkId[message.Filename]; ok {
		tempChunkList = append(tempChunkList, val...)
	}
	metaData.fileIdToChunkId[message.Filename] = append(tempChunkList, chunkId)
	lock.Unlock()

	// ask 3 chunkserver to create chunks
	new_chunkServer := choose_n_random_chunkServers()
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
	helper.SendMessage(message1)

	// increment offset
	lock.Lock()
	metaData.chunkIdToOffset[chunkId] += message1.PayloadSize
	lock.Unlock()
}

// after receiving ack from primary, approve request for client
func ackChunkCreate(message structs.Message) {
	messagePorts := message.Ports // [C, M, P, S1, S2]
	messagePorts = append([]int{messagePorts[0]}, messagePorts[2:]...) //[C, P, S1, S2]
	chunkServers := messagePorts[1:]

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
	helper.SendMessage(message1)

	//record in metaData
	lock.Lock()
	for _, v := range chunkServers {
		metaData.chunkServerToChunkId[v] = append(metaData.chunkServerToChunkId[v], message.ChunkId)
	}
	metaData.chunkIdToChunkserver[message.ChunkId] = chunkServers

	// increment offset
	// new_offset := metaData.chunkIdToOffset[message.ChunkId] + message1.PayloadSize
	// metaData.chunkIdToOffset[message.ChunkId] = new_offset
	lock.Unlock()
}

func (m heartState) String() string {
	return [...]string{"START", "PENDING", "ALIVE", "DEAD"}[m]
}

// Separate go-routine to send heartbeat messages in intervals
func sendHeartbeat() {

	for {
		for _, liveServer := range liveChunkServers {
			currentHeartState, _ := metaData.heartBeatAck.Load(liveServer)
			switch currentHeartState {
			case Start: // For the first heartbeat
				metaData.heartBeatAck.Store(liveServer, Pending)
				heartbeatMsg := structs.Message{
					MessageType: helper.HEARTBEAT,
					Ports:       []int{helper.MASTER_SERVER_PORT, liveServer}, // [C, P, S1, S2]
					Pointer:     1,
				}
				go helper.SendMessage(heartbeatMsg)
			case Pending: // No reply from the previous heartbeat, consider the chunk server dead
				go startReplicate(liveServer) // give the port number to function
				metaData.heartBeatAck.Store(liveServer, Dead)
			case Alive: // Successfully received ack from previous heartbeat, send next heartbeat
				metaData.heartBeatAck.Store(liveServer, Pending)
				heartbeatMsg := structs.Message{
					MessageType: helper.HEARTBEAT,
					Ports:       []int{helper.MASTER_SERVER_PORT, liveServer}, // [C, P, S1, S2]
					Pointer:     1,
				}
				go helper.SendMessage(heartbeatMsg)
			case Dead: // Chunk server is dead, replicate server
				fmt.Println(liveServer, " is dead!")
			}

		}

		time.Sleep(helper.HEARTBEAT_TIMEOUT)
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
					metaData.chunkIdToChunkserver[failedChunkId] = append(metaData.chunkIdToChunkserver[failedChunkId][:idx], metaData.chunkIdToChunkserver[failedChunkId][idx+1:]...)
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
	for _, replicationMsgs := range temporaryReplicationMap {
		for _, repMessage := range replicationMsgs {
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

func choose_n_random_chunkServers() []int {

	chunkServerArray := make(map[int]bool)
	for _, server := range liveChunkServers {
		serverStatus, _ := metaData.heartBeatAck.Load(server)
		if serverStatus == Alive {
			chunkServerArray[server] = false
		}
	}
	res := []int{}

	for len(res) < helper.REP_COUNT {
		//random key stores the key from the chunkS
		random_key := MapRandomKeyGet(chunkServerArray).(int)

		if len(metaData.chunkServerToChunkId[random_key]) >= helper.CHUNK_CAPACITY {
			chunkServerArray[random_key] = true
			if !Contains(fullCapacityServers, random_key) {
				fullCapacityServers = append(fullCapacityServers, random_key)
				fmt.Println("A new chunk server with port ", latestChunkServer, " has been spun up as chunk server ", random_key, " is full.")
				spinUpNewCS(len(liveChunkServers) - 1)
			}
		}

		// checking if this key boolean is false or true, if false append this key to the res and set the key value true instead
		if chunkServerArray[random_key] == false {
			chunkServerArray[random_key] = true
			res = append(res, random_key)
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
func (m *MetaData) initialiseACKMap(portNum int) {
	m.heartBeatAck.Store(portNum, Start)
}

func spinUpNewCS(i int) {
	go chunkServer.ChunkServer(i, latestChunkServer)
	metaData.initialiseACKMap(latestChunkServer)
	liveChunkServers = append(liveChunkServers, latestChunkServer)
	latestChunkServer += 1
}

func main() {
	metaData.fileIdToChunkId = make(map[string][]string)
	metaData.chunkIdToChunkserver = make(map[string][]int)
	metaData.chunkServerToChunkId = make(map[int][]string)
	metaData.chunkIdToOffset = make(map[string]int64)
	metaData.replicationMap = make(map[int][]structs.RepMsg)
	metaData.successfulWrites = make(map[string][]structs.SuccessfulWrite)
	//metaData.initialiseACKMap()
	for i := 0; i < helper.INITIAL_NUM_CHUNKSERVERS; i++ {
		spinUpNewCS(i)		
	}

	go sendHeartbeat()
	listen(0, helper.MASTER_SERVER_PORT)

}