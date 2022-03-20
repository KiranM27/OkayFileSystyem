package main

import (
	"sync"
	"github.com/gin-gonic/gin"
	helper "oks/Helper"
	structs "oks/Structs"
	"net/http"
	"fmt"
	"strconv"
)

var (
	buffer sync.Map 
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
	context.IndentedJSON(http.StatusOK, message.MessageType + " Received")
	fmt.Println("---------- Received Message: ", message.MessageType, " ----------")

	switch message.MessageType {
	case helper.DATA_APPEND:
		appendMessageHandler(message)
	case helper.ACK_APPEND:
		appendACKHandler(message)
	// case helper.DATA_COMMIT:
	// 	commitDataHandler(message)
	// case helper.ACK_COMMIT:
	// 	commitACKHandler(message)
	// case helper.CREATE_NEW_CHUNK:
	// 	createNewChunkHandler(message)
	// case helper.DATA_PAD:
	// 	padHandler(message)
	// case helper.ACK_PAD:
	// 	padACKHandler(message)
	}
}

func appendMessageHandler(message structs.Message) {
	buffer.Store(message.GenerateUid(), message.Payload)
	if message.Pointer == len(message.Ports) - 1 {
		message.Reply()
		message.SetMessageType(helper.ACK_APPEND)
	} else {
		message.Forward()
	}
	helper.SendMessage(message)
}

func appendACKHandler(message structs.Message) {
	if message.Pointer != 0 {
		message.Reply()
		helper.SendMessage(message)
	}
}

func listen(nodePid int, portNo int) {
	router := gin.Default()
	router.GET("/", landingPageHandler)
	router.POST("/message", postMessageHandler)

	fmt.Printf("Node %d listening on port %d \n", nodePid, portNo)
	router.Run("localhost:" + strconv.Itoa(portNo))
}

func main() {
	
	go listen(1, 8001)
	go listen(2, 8002)
	go listen(3, 8003)

	message := structs.Message{
		MessageType: helper.DATA_APPEND,
		Ports:       []int{8080, 8001, 8002, 8003}, // 0: Client, 1: Primary, 2+: Secondary
		Pointer:     1,
		Filename:    "hola",
		ChunkId:     "hello",
		Payload:     "paylaod",
		PayloadSize: 7,
		ChunkOffset: 0,
	}

	helper.SendMessage(message)
	for {}
}