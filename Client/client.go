package client

import (
	//"os"
	"fmt"
	"net/http"
	helper "oks/Helper"
	structs "oks/Structs"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func listen(id int, portNumber int) {
	router := gin.Default()
	router.POST("/message", messageHandler)

	fmt.Printf("Client %d listening on port %d \n", id, portNumber)
	router.Run("localhost:" + strconv.Itoa(portNumber))
}

func messageHandler(context *gin.Context) {

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
		go sendChunkAppend(message, true, context)
	case helper.ACK_APPEND:
		fmt.Println("Chunk gave a reply for append request")
		go confirmWrite(message, true, context)
	}
}

// Send a request to Master that client wants to append
// The message has the size of the data
func requestMasterAppend(clientPort int, sourceFilename string, OFSFilename string) {
	var numChunks uint64

	// Call helper function to read source file size
	sourceFileByteSize := getFileSize(sourceFilename)

	// Check byte size of file, if more than 2.5kb split
	if sourceFileByteSize > 2500 {
		numChunks = splitFile(sourceFilename)
	} else {
		numChunks = 1
	}

	// If no split happened, append normally
	// If split, for loop to append
	if numChunks == 1 {
		message := structs.Message{
			MessageType:    helper.DATA_APPEND,
			Ports:          []int{clientPort, helper.MASTER_SERVER_PORT}, // 0 is client, 1 is master
			Pointer:        1,
			SourceFilename: sourceFilename,
			Filename:       OFSFilename, // File name in the Okay File System
			PayloadSize:    sourceFileByteSize,
		}
		// HTTP Request to Master
		fmt.Println(message)
		fmt.Println("Sending append request to Master")
		helper.SendMessage(message)
		fmt.Println("Finished sending append request to Master")

	} else {
		sourceFilePrefix := removeExtension(sourceFilename)
		for i := uint64(0); i < numChunks; i++ {
			smallFileName := sourceFilePrefix + strconv.FormatUint(i, 10) + ".txt"
			smallFileSize := getFileSize(smallFileName)
			fmt.Println(smallFileName) // debug
			fmt.Println(smallFileSize) // debug

			message := structs.Message{
				MessageType:    helper.DATA_APPEND,
				Ports:          []int{clientPort, helper.MASTER_SERVER_PORT}, // 0 is client, 1 is master
				Pointer:        1,
				SourceFilename: smallFileName,
				Filename:       OFSFilename, // TODO: This should be the same file name still right
				PayloadSize:    smallFileSize,
			}
			fmt.Println(message)
			// HTTP Request to Master
			fmt.Println("Sending append request to Master")
			helper.SendMessage(message)
		}
	}
}

// // Check post route for the message along with timeout
// func timeoutCheck(context *gin.Context)bool{
// 	var message structs.Message

// 	// Call BindJSON to bind the received JSON to message.
// 	if err := context.BindJSON(&message); err != nil {
// 		fmt.Println("Invalid message object received.")
// 		return false
// 	}
// 	context.IndentedJSON(http.StatusOK, "Client has received message")

// 	switch message.MessageType {
// 	case helper.ACK_APPEND:
// 		fmt.Println("Chunk gave a reply for append ack request")
// 		sendChunkAppend(message, true, context)
// 	case helper.ACK_COMMIT:
// 		fmt.Println("Chunk gave a reply for commit ack request")
// 		confirmWrite(message, true, context)
// 	}

// }

// Send append request to primary chunk server and wait
func sendChunkAppend(message structs.Message, tryAgain bool, context *gin.Context) {
	fmt.Println("Processing append request to primary chunk server")

	// message.Pointer += 1 // increment the pointer first (initial pointer should be 0)
	message.Forward()
	message.Payload = readFile(message.SourceFilename)

	// HTTP Request to Primary Chunk
	fmt.Println("Sending append request to Primary Chunk Server")
	fmt.Println(message.Pointer, message.Ports)
	//success := SendTimerMessage(message)
	helper.SendMessage(message)

	// Check Post Route for response with a timeout
	// success := timeoutCheck(context)
	success := true

	// If timed out and can try again, run sendChunkAppend again
	if !success && tryAgain {
		//message.Pointer -= 1 // decrement the pointer
		message.Reply()
		sendChunkAppend(message, false, context)
	} else if !success && !tryAgain {
		// if failed and cannot try again,
		fmt.Println("Append has failed, restart entire process later")
		time.Sleep(time.Second * 30) // simulate pause
		requestMasterAppend(message.Ports[0], message.SourceFilename, message.Filename)

	} else { // success either case
		fmt.Println("Append succeeded, proceed to confirm write")
		// there should be a function here
	}
}

// Confirm write to the chunk servers
func confirmWrite(message structs.Message, tryAgain bool, context *gin.Context) {
	fmt.Println("Confirming write request to primary chunk server")
	// message.Pointer += 1 // increment the pointer first (initial pointer should be 0)
	message.Forward()
	message.MessageType = helper.DATA_COMMIT

	// HTTP Request to Primary Chunk
	fmt.Println("Sending append request to Primary Chunk Server")
	helper.SendMessage(message)

	// Check Post Route for response with a timeout
	//success := timeoutCheck(context)
	success := true

	// If timed out and can try again, run confirmWrite again
	if !success && tryAgain {
		//message.Pointer -= 1 // decrement the pointer
		message.Reply()
		confirmWrite(message, false, context)
	} else if !success && !tryAgain {
		// if failed and cannot try again,
		fmt.Println("Write has failed, restart entire process later")
		time.Sleep(time.Second * 30) // simulate pause
		requestMasterAppend(message.Ports[0], message.SourceFilename, message.Filename)

	} else { // success either case
		fmt.Println("Write succeeded, Client successfully appended")
	}
}

func InitClient(id int, portNumber int, sourceFilename string, OFSFilename string) {
	fmt.Printf("Client %d is going up at %d\n", id, portNumber)
	go listen(id, portNumber)
	requestMasterAppend(portNumber, sourceFilename, OFSFilename)
	for {

	}
}
