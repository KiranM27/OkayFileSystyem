package client

import (
	//"os"
	"time"
	"fmt"
	"strconv"
	"net/http"
	"github.com/gin-gonic/gin"
	"sync"
	structs "oks/Structs"
	helper "oks/Helper"
)

var ACKMap sync.Map

func listen(id int, portNumber int){
	router := gin.Default()
	router.POST("/message", messageHandler)

	fmt.Printf("Client %d listening on port %d \n", id, portNumber)
	router.Run("localhost:" + strconv.Itoa(portNumber))
}

func messageHandler(context *gin.Context){

	var message structs.Message
	//var replyHere chan[]bool
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
	case helper.ACK_COMMIT:
		fmt.Println("Chunk gave reply for commit request")
		go finishAppend(message,context)
	}
}
// Send a request to Master that client wants to append
func requestMasterAppend(clientPort int, filename string) {
	var numChunks uint64
	var message structs.Message
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
		message = structs.Message{
			MessageType: helper.DATA_APPEND, 
			Ports: []int{clientPort, helper.MASTER_SERVER_PORT}, // 0 is client, 1 is master
			Pointer: 1,
			Filename: filename,
			PayloadSize: fileByteSize,
			RecordIndex: uint64(0),
		}
		ACKMap.Store(int(message.RecordIndex),true)
		// HTTP Request to Master
		fmt.Println("Sending append request to Master")
		helper.SendMessage(message)
		go runTimer(message)
	} else{
		filePrefix := removeExtension(filename)
		for i := uint64(0); i < numChunks; i++{
			smallFileName := filePrefix + strconv.FormatUint(i, 10) + ".txt"
			smallFileSize := getFileSize(smallFileName)
			fmt.Println(smallFileName) // debug
			fmt.Println(smallFileSize) // debug

			message = structs.Message{
				MessageType: helper.DATA_APPEND, 
				Ports: []int{clientPort, helper.MASTER_SERVER_PORT}, // 0 is client, 1 is master
				Pointer: 1,
				Filename: smallFileName,
				PayloadSize: smallFileSize,
				RecordIndex: i,
			}
			ACKMap.Store(int(message.RecordIndex),true)
			// HTTP Request to Master
			fmt.Println("Sending append request to Master")
			helper.SendMessage(message)
			go runTimer(message)
		}
		
	}
}

func runTimer(message structs.Message){
	timer := time.NewTimer(3 * time.Second) 
	for {
		timer.Reset(5 *  time.Second) 
		select{
		case <- timer.C:
			_, locked:= ACKMap.Load(int(message.RecordIndex))
			fmt.Println(locked)
			if locked{
				fmt.Println("Timeout, resending message")
				helper.SendMessage(message)
			} else {
				return
			}
		}
	}
}



// Send append request to primary chunk server and wait
func sendChunkAppend(message structs.Message, tryAgain bool, context *gin.Context){
	fmt.Println("Processing append request to primary chunk server") 

	// message.Pointer += 1 // increment the pointer first (initial pointer should be 0)
	message.Forward()
	message.Payload = readFile(message.Filename)

	// HTTP Request to Primary Chunk
	fmt.Println("Sending append request to Primary Chunk Server")
	//success := SendTimerMessage(message)
	helper.SendMessage(message)

}
// Confirm write to the chunk servers
func confirmWrite(message structs.Message,	tryAgain bool, context *gin.Context){
	fmt.Println("Confirming write request to primary chunk server") 
	// message.Pointer += 1 // increment the pointer first (initial pointer should be 0)
	message.Forward()
	message.MessageType = helper.DATA_COMMIT

	// HTTP Request to Primary Chunk
	fmt.Println("Sending append request to Primary Chunk Server")
	helper.SendMessage(message)
}

func finishAppend(message structs.Message, context *gin.Context){
	ACKMap.Delete(message.RecordIndex)
}

func InitClient(id int,portNumber int){
	fmt.Printf("Client %d is going up at %d\n", id, portNumber)
	go listen(id, portNumber)
	requestMasterAppend(portNumber, "test.txt")
	for{

	}
}
