package chunkServer

import (
	"errors"
	"fmt"
	"net/http"
	helper "oks/Helper"
	structs "oks/Structs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"github.com/gin-gonic/gin"
)

var (
	buffer   sync.Map
	mutex    sync.Mutex
	aliveMap sync.Map
)

func landingPageHandler(context *gin.Context) {
	context.IndentedJSON(http.StatusOK, "Welcome to the Okay File System ! This is a chunk server.")
}

func postMessageHandler(context *gin.Context) {
	var message structs.Message

	// Call BindJSON to bind the received JSON to message.
	if err := context.BindJSON(&message); err != nil {
		fmt.Println("Invalid message object received.", err)
		return
	}
	context.IndentedJSON(http.StatusOK, message.MessageType+" Received")
	fmt.Println("---------- Received Message: ", message.MessageType, " ----------")

	portNo := strconv.Itoa(message.Ports[message.Pointer])
	isNodealive, _ := aliveMap.Load(portNo)
	if isNodealive == false && message.MessageType != helper.REVIVE {
		fmt.Println("Node " + portNo + " is dead and will not be responding to the incoming request.")
		return
	}

	switch message.MessageType {
	case helper.DATA_APPEND:
		go appendMessageHandler(message)
	case helper.ACK_APPEND:
		go ACKHandler(message)
	case helper.DATA_COMMIT:
		go commitDataHandler(message)
	case helper.ACK_COMMIT:
		go ACKHandler(message)
	case helper.CREATE_NEW_CHUNK:
		createNewChunkHandler(message)
	case helper.ACK_CHUNK_CREATE:
		go ACKHandler(message)
	case helper.HEARTBEAT:
		go heartbeatHandler(message)
	case helper.KILL_YOURSELF:
		go killYourselfHandler(message)
	case helper.REVIVE:
		go reviveHandler(message)
	}
}

func repMessageHandler(context *gin.Context) {
	var repMsg structs.RepMsg
	portNo := strconv.Itoa(repMsg.TargetCS)

	// Call BindJSON to bind the received JSON to message.
	if err := context.BindJSON(&repMsg); err != nil {
		fmt.Println("Invalid message object received.", err)
		return
	}
	context.IndentedJSON(http.StatusOK, repMsg.MessageType+" Received")
	fmt.Println("---------- Received Replication Message: ", repMsg.MessageType, " ----------")

	isNodealive, _ := aliveMap.Load(portNo)
	if isNodealive == false {
		fmt.Println("Node " + portNo + " is dead and will not be responding to the incoming request.")
		return
	}

	switch repMsg.MessageType {
	case helper.REPLICATE:
		go replicateHandler(repMsg, 0)
	case helper.REP_DATA_REQUEST:
		go repDataRequestHandler(repMsg)
	case helper.REP_DATA_REPLY:
		go repDataReplyHandler(repMsg)
		// Take data from call, create a new file, write data and then send ACK_REPPLICAION to master
	}
}

func readMsgHandler(context *gin.Context) {
	var readMsg structs.ReadMsg

	// Call BindJSON to bind the received JSON to message.
	if err := context.BindJSON(&readMsg); err != nil {
		fmt.Println("Invalid message object received.", err)
		return
	}
	portNo := strconv.Itoa(readMsg.Sources[0])
	isNodealive, _ := aliveMap.Load(portNo)
	if isNodealive == false {
		fmt.Println("Node " + portNo + " is dead and will not be responding to the incoming request.")
		return
	}

	content := readReqHandler(readMsg)
	context.IndentedJSON(http.StatusOK, content)
}

func ACKHandler(message structs.Message) {
	if message.Pointer != 0 {
		message.Reply()
		helper.SendMessage(message)
	}
}

func appendMessageHandler(message structs.Message) {
	buffer.Store(message.GenerateUid(), message.Payload)
	if message.Pointer == len(message.Ports)-1 {
		message.Reply()
		message.SetMessageType(helper.ACK_APPEND)
	} else {
		message.Forward()
	}
	helper.SendMessage(message)
}

func commitDataHandler(message structs.Message) {
	mutex.Lock()
	err := writeMutation(message.ChunkId, message.ChunkOffset, message.GenerateUid(), message.Ports[message.Pointer])
	if err != nil {
		fmt.Println(err)
	} else {
		if message.Pointer == len(message.Ports)-1 {
			message.Reply()
			message.SetMessageType(helper.ACK_COMMIT)
		} else {
			message.Forward()
		}
		helper.SendMessage(message)
	}
	mutex.Unlock()
}

func createNewChunkHandler(message structs.Message) {
	createChunk(message.Ports[message.Pointer], message.ChunkId)
	if message.Pointer == len(message.Ports)-1 {
		message.Reply()
		message.SetMessageType(helper.ACK_CHUNK_CREATE)
	} else {
		message.Forward()
	}
	helper.SendMessage(message)
}

func heartbeatHandler(message structs.Message) {
	message.SetMessageType(helper.ACK_HEARTBEAT)
	message.Reply()
	helper.SendMessage(message)
}

func killYourselfHandler(message structs.Message) {
	portNo := strconv.Itoa(message.Ports[message.Pointer])
	fmt.Println("Kill message received by Node " + portNo)
	dataPath := ".." + helper.DATA_DIR + "/" + portNo
	absDataPath, _ := filepath.Abs(dataPath)
	err := os.RemoveAll(absDataPath)
	if err != nil {
		fmt.Println("Error occurred while clearing files in port: "+portNo, err)
	}

	aliveMap.Store(portNo, false) // Set isAlive to False.
	fmt.Println("Node " + portNo + " has failed!")
}

func reviveHandler(message structs.Message) {
	portNo := strconv.Itoa(message.Ports[message.Pointer])
	aliveMap.Store(portNo, true)
	message.SetMessageType(helper.ACK_HEARTBEAT)
	message.Reply()
	helper.SendMessage(message)
	fmt.Println("Node " + portNo + " is now back up!")
}

func replicateHandler(repMsg structs.RepMsg, chunkServerIdx int) {
	fmt.Println("YOOOOOOOOOOOOOOOOOOOO ", repMsg.TargetCS)
	time.Sleep(time.Second * 5)
	repMsg.SetMessageType(helper.REP_DATA_REQUEST)
	fmt.Println("____________________ Source: ", repMsg.Sources, "________________________")
	helper.SendRepMsg(repMsg, repMsg.Sources[chunkServerIdx])
	// ACKMap.Store(repMsg.TargetCS, )
	// go runTimer(repMsg, chunkServerIdx)
}

func repDataRequestHandler(repMsg structs.RepMsg) {
	chunkId := repMsg.ChunkId
	chunkSourcePort := repMsg.Sources[0]
	pwd, _ := os.Getwd()
	dataDirPath := filepath.Join(pwd, "../"+helper.DATA_DIR)
	portDirPath := filepath.Join(dataDirPath, strconv.Itoa(chunkSourcePort))
	chunkPath := filepath.Join(portDirPath, chunkId+".txt")
	content := helper.ReadFile(chunkPath)
	fmt.Println("")
	repMsg.SetMessageType(helper.REP_DATA_REPLY)
	repMsg.SetPayload(content)
	helper.SendRepMsg(repMsg, repMsg.TargetCS)
}

func repDataReplyHandler(repMsg structs.RepMsg) error {
	chunkId := repMsg.ChunkId
	currentPort := repMsg.TargetCS
	pwd, _ := os.Getwd()
	dataDirPath := filepath.Join(pwd, "../"+helper.DATA_DIR)
	portDirPath := filepath.Join(dataDirPath, strconv.Itoa(currentPort))
	chunkPath := filepath.Join(portDirPath, chunkId+".txt")

	createChunk(repMsg.TargetCS, chunkId)
	fh, err := os.OpenFile(chunkPath, os.O_WRONLY, 0777)
	if err != nil {
		fmt.Println(err)
	}

	defer fh.Close()
	// write data
	writeDataBytes := []byte(repMsg.Payload)
	if _, err := fh.Write(writeDataBytes); err != nil {
		fmt.Println(err)
		return errors.New("Write Failed for Replication of chunk " + chunkId)
	}

	repMsg.SetMessageType(helper.ACK_REPLICATION)
	helper.SendRepMsg(repMsg, helper.MASTER_SERVER_PORT)
	return nil
}

func readReqHandler(readMsg structs.ReadMsg) string {
	chunkId := readMsg.ChunkId
	chunkSourcePort := readMsg.Sources[0]
	pwd, _ := os.Getwd()
	dataDirPath := filepath.Join(pwd, "../"+helper.DATA_DIR)
	portDirPath := filepath.Join(dataDirPath, strconv.Itoa(chunkSourcePort))
	chunkPath := filepath.Join(portDirPath, chunkId+".txt")
	content := helper.ReadFile(chunkPath)
	filteredOutput := filterContentBySW(content, readMsg.SuccessfulWrites)
	return filteredOutput
}

func writeMutation(chunkId string, chunkOffset int64, uid string, currentPort int) error {
	pwd, _ := os.Getwd()
	dataDirPath := filepath.Join(pwd, "../"+helper.DATA_DIR)
	portDirPath := filepath.Join(dataDirPath, strconv.Itoa(currentPort))
	chunkPath := filepath.Join(portDirPath, chunkId+".txt")
	fh, err := os.OpenFile(chunkPath, os.O_WRONLY, 0777)
	if err != nil {
		fmt.Println(err)
	}

	defer fh.Close()
	// write data
	writeData, ok := buffer.Load(uid)

	if !ok {
		return errors.New("No data in Buffer for UID " + uid)
	}

	fileStats, _ := fh.Stat()
	fileSize := fileStats.Size()
	if chunkOffset > fileSize {
		writeData = strings.Repeat(helper.PADDING, int(chunkOffset)-int(fileSize)) + writeData.(string)
	}

	writeDataBytes := []byte(writeData.(string))
	_, err = fh.Seek(fileSize, 0)
	if err != nil {
		return errors.New("Error while seeking to UID " + uid)
	}

	if _, err := fh.Write(writeDataBytes); err != nil {
		fmt.Println(err)
		return errors.New("Write Failed for UID " + uid)
	}
	return nil
}

func createChunk(portNo int, chunkId string) {
	pwd, _ := os.Getwd()
	dataDirPath := filepath.Join(pwd, "../"+helper.DATA_DIR)
	helper.CreateFolder(dataDirPath)
	portDataDirPath := filepath.Join(dataDirPath, strconv.Itoa(portNo))
	helper.CreateFolder(portDataDirPath)
	chunkPath := filepath.Join(portDataDirPath, chunkId+".txt")
	helper.CreateFile(chunkPath)
}

func filterContentBySW(content string, successfulWrites []structs.SuccessfulWrite) string {
	_content := []byte(content)
	var output []byte
	for _, SW := range successfulWrites {
		output = append(output, _content[SW.Start:SW.End]...)
	}
	filteredOutput := string(output)
	return filteredOutput
}


func listen(nodePid int, portNo int) {

	router := gin.Default()

	router.GET("/", landingPageHandler)
	router.POST("/message", postMessageHandler)
	router.POST("/replicate", repMessageHandler)
	router.POST("/read", readMsgHandler)

	fmt.Printf("Node %d listening on port %d \n", nodePid, portNo)
	router.Run("localhost:" + strconv.Itoa(portNo))
}

func ChunkServer(nodePid int, portNo int) {
	go listen(nodePid, portNo)
}

// var ACKMap sync.Map

// func runTimer(repMsg structs.RepMsg,chunkServerIdx int){
// 	t := time.After(100 * time.Second)
// 	select{
// 	case <- t:
// 		if
// 		fmt.Printf("%v: Replication Timeout, trying next server", repMsg.TargetCS)
// 	}
// }
