package client

import (
	//"os"
	"time"
	"fmt"
	"strconv"
	"net/http"
	"github.com/gin-gonic/gin"
	structs "oks/Structs"
	helper "oks/Helper"
)

func listen(id int, portNumber int){
	router := gin.Default()
	router.POST("/message", messageHandler)

	fmt.Printf("Client %d listening on port %d \n", id, portNumber)
	router.Run("localhost:" + strconv.Itoa(portNumber))
}

func messageHandler(context *gin.Context){

	var message structs.Message

	// Call BindJSON to bind the received JSON to message.
	if err := context.BindJSON(&message); err != nil {
		fmt.Println("Invalid message object received.")
		return
	}
	context.IndentedJSON(http.StatusOK, "New placeholder")

	switch message.MessageType {
	case helper.DATA_APPEND:
		fmt.Println("Master gave a reply for append request")
		go sendChunkAppend(message, true)
	case helper.ACK_APPEND:
		fmt.Println("Chunk gave a reply for append request")
		go confirmWrite(message, true)
	}
}
// Send a request to Master that client wants to append
func requestMasterAppend(clientPort int, filename string) {
	var numChunks uint64

	// Call helper function to read file size
	fileByteSize := getFileSize(filename)

	// Check byte size of file, if more than 2.5kb split
	if(fileByteSize > 2500){
		numChunks = splitFile(filename)
	} else{
		numChunks = 1
	}

	// If no split happened, append normally
	// If split, for loop to append
	if numChunks == 1{
		message := structs.Message{
			MessageType: helper.DATA_APPEND, 
			Ports: []int{clientPort, helper.MASTER_SERVER_PORT}, // 0 is client, 1 is master
			Pointer: 1,
			Filename: filename,
			PayloadSize: fileByteSize,
	
		}
		// HTTP Request to Master
		fmt.Println(message)
		fmt.Println("Sending append request to Master")
		helper.SendMessage(message)
	} else{
		filePrefix := removeExtension(filename)
		for i := uint64(0); i < numChunks; i++{
			smallFileName := filePrefix + strconv.FormatUint(i, 10) + ".txt"
			smallFileSize := getFileSize(smallFileName)
			fmt.Println(smallFileName) // debug
			fmt.Println(smallFileSize) // debug

			message := structs.Message{
				MessageType: helper.DATA_APPEND, 
				Ports: []int{clientPort, helper.MASTER_SERVER_PORT}, // 0 is client, 1 is master
				Pointer: 1,
				Filename: smallFileName,
				PayloadSize: smallFileSize,
		
			}
			fmt.Println(message)
			// HTTP Request to Master
			fmt.Println("Sending append request to Master")
			helper.SendMessage(message)
		}
	}
}

// Send append request to primary chunk server and wait
func sendChunkAppend(message structs.Message, tryAgain bool){
	fmt.Println("Processing append request to primary chunk server") 

	message.Pointer += 1 // increment the pointer first (initial pointer should be 0)
	message.Payload = readFile(message.Filename)

	// HTTP Request to Primary Chunk
	fmt.Println("Sending append request to Primary Chunk Server")
	success := SendTimerMessage(message)

	// If timed out and can try again, run sendChunkAppend again
	if (!success && tryAgain){
		message.Pointer -= 1 // decrement the pointer
		sendChunkAppend(message, false)
	} else if (!success && !tryAgain){ 
		// if failed and cannot try again,
		fmt.Println("Append has failed, restart entire process later")
		time.Sleep(time.Second * 30) // simulate pause
		requestMasterAppend(message.Ports[0], message.Filename)

	} else{ // success either case
		fmt.Println("Append succeeded, proceed to confirm write")
		// there should be a function here
	}
}
// Confirm write to the chunk servers
func confirmWrite(message structs.Message,	tryAgain bool){
	fmt.Println("Confirming write request to primary chunk server") 
	message.Pointer += 1 // increment the pointer first (initial pointer should be 0)
	message.MessageType = helper.DATA_COMMIT

	// HTTP Request to Primary Chunk
	fmt.Println("Sending append request to Primary Chunk Server")
	success := SendTimerMessage(message)

	// If timed out and can try again, run confirmWrite again
	if (!success && tryAgain){
		message.Pointer -= 1 // decrement the pointer
		confirmWrite(message, false)
	} else if (!success && !tryAgain){ 
		// if failed and cannot try again,
		fmt.Println("Write has failed, restart entire process later")
		time.Sleep(time.Second * 30) // simulate pause
		requestMasterAppend(message.Ports[0], message.Filename)

	} else{ // success either case
		fmt.Println("Write succeeded, Client successfully appended")
	}
}


func InitClient(id int,portNumber int){
	fmt.Printf("Client %d is going up at %d\n", id, portNumber)
	go listen(id, portNumber)
	requestMasterAppend(portNumber, "test.txt")
}