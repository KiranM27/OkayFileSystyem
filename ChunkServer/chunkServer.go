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

	"github.com/gin-gonic/gin"
)

var (
	buffer sync.Map
	mutex  sync.Mutex
)

func landingPageHandler(context *gin.Context) {
	context.IndentedJSON(http.StatusOK, "Welcome to the Okay File System ! This is a chunk server.")
}

func postMessageHandler(context *gin.Context) {
	var message structs.Message

	// Call BindJSON to bind the received JSON to message.
	if err := context.BindJSON(&message); err != nil {
		fmt.Println("Invalid message object received.")
		return
	}
	context.IndentedJSON(http.StatusOK, message.MessageType+" Received")
	fmt.Println("---------- Received Message: ", message.MessageType, " ----------")

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
		// case helper.DATA_PAD:
		// 	padHandler(message)
		// case helper.ACK_PAD:
		// 	padACKHandler(message)
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

func ACKHandler(message structs.Message) {
	if message.Pointer != 0 {
		message.Reply()
		helper.SendMessage(message)
	}
}

func commitDataHandler(message structs.Message) {
	mutex.Lock()
	err := writeMutation(message.ChunkId, message.ChunkOffset, message.GenerateUid(), message.Ports[message.Pointer])
	//fmt.Println("ERROR :", err)
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

func writeMutation(chunkId string, chunkOffset int64, uid string, currentPort int) error {
	pwd, _ := os.Getwd()
	dataDirPath := filepath.Join(pwd, "../"+helper.DATA_DIR)
	portDirPath := filepath.Join(dataDirPath, strconv.Itoa(currentPort))
	chunkPath := filepath.Join(portDirPath, chunkId+".txt")
	fmt.Println(chunkPath)
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
	_, err = fh.Seek(chunkOffset, 0)
	if err != nil {
		return errors.New("Error while seeking to UID " + uid)
	}
	//fmt.Println(writeDataBytes)
	if _, err := fh.Write(writeDataBytes); err != nil {
		fmt.Println(err)
		return errors.New("Write Failed for UID " + uid)
	}
	return nil
}

func listen(nodePid int, portNo int) {
	router := gin.Default()

	router.GET("/", landingPageHandler)
	router.POST("/message", postMessageHandler)

	fmt.Printf("Node %d listening on port %d \n", nodePid, portNo)
	router.Run("localhost:" + strconv.Itoa(portNo))
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

func testCreateChunk() {
	createChunk(8081, "test_c0")
	createChunk(8082, "test_c0")
	createChunk(8083, "test_c0")
}

func ChunkServer(nodePid int, portNo int) {
	go listen(nodePid, portNo)
}


	
// func main() {

// go listen(1, 8000)
// go listen(2, 8002)
// go listen(3, 8003)

// buffer.Store("holahello9", "fuckgo")

// message := structs.Message{
// 	MessageType: helper.DATA_COMMIT,
// 	Ports:       []int{8080, 8000, 8002, 8003}, // 0: Client, 1: Primary, 2+: Secondary
// 	Pointer:     1,
// 	Filename:    "hola",
// 	ChunkId:     "hello",
// 	Payload:     "fuckgo",
// 	PayloadSize: 8,
// 	ChunkOffset: 9,
// }

// helper.SendMessage(message)

// for {
// }
// testCreateChunk()

// }
