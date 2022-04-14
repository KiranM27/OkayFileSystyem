package client

// Client works on only one append operration at a time !!

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	helper "oks/Helper"
	structs "oks/Structs"
	"strconv"
	"sync"
	"time"
)

var ACKMap sync.Map

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
		go sendChunkAppend(message)
	case helper.ACK_APPEND:
		go confirmWrite(message)
	case helper.ACK_COMMIT:
		go finishAppend(message)
	}
}

func readChunk(chunkId string) {

}

// Send a request to Master that client wants to append
// The message has the size of the data
func requestMasterAppend(clientPort int, sourceFilename string, OFSFilename string) {
	var numChunks uint64
	//var message structs.Message

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
		ACKMapClientRecords, _ := ACKMap.Load(clientPort)
		if ACKMapClientRecords == nil {
			ACKMapClientRecords = []structs.ACKMAPRecord{}
		}
		recordIndex := len(ACKMapClientRecords.([]structs.ACKMAPRecord))
		ACKMapClientRecords = append(ACKMapClientRecords.([]structs.ACKMAPRecord), structs.ACKMAPRecord{
			RecordIndex: recordIndex,
			Acked:       false,
		})
		message := structs.Message{
			MessageType:    helper.DATA_APPEND,
			Ports:          []int{clientPort, helper.MASTER_SERVER_PORT}, // 0 is client, 1 is master
			Pointer:        1,
			SourceFilename: sourceFilename,
			Filename:       OFSFilename, // File name in the Okay File System
			PayloadSize:    sourceFileByteSize,
			RecordIndex:    uint64(recordIndex),
		}

		ACKMap.Store(clientPort, ACKMapClientRecords)
		// HTTP Request to Master
		fmt.Println(strconv.Itoa(clientPort) + " Sending append request to Master")
		helper.SendMessage(message)
		go runTimer(message)
		// fmt.Println(strconv.Itoa(clientPort) + " Finished sending append request to Master")

	} else {
		// TODO: FIX RECORD INDEX - USE SOLN ABOVE, BUT NEED TO UPDATE FOR MULTIPLE RECORDS
		sourceFilePrefix := removeExtension(sourceFilename)
		for i := uint64(0); i < numChunks; i++ {
			smallFileName := sourceFilePrefix + strconv.FormatUint(i, 10) + ".txt"
			smallFileSize := getFileSize(smallFileName)
			// fmt.Println(smallFileName) // debug
			// fmt.Println(smallFileSize) // debug

			message := structs.Message{
				MessageType:    helper.DATA_APPEND,
				Ports:          []int{clientPort, helper.MASTER_SERVER_PORT}, // 0 is client, 1 is master
				Pointer:        1,
				SourceFilename: smallFileName,
				Filename:       OFSFilename, // TODO: This should be the same file name still right
				PayloadSize:    smallFileSize,
			}
			ACKMap.Store(int(message.RecordIndex), true)
			// HTTP Request to Master
			fmt.Println(strconv.Itoa(clientPort) + " Sending append request to Master")
			helper.SendMessage(message)
			go runTimer(message)
		}

	}
}

func runTimer(message structs.Message) {
	timer := time.NewTimer(3 * time.Second)
	clientPort := message.Ports[0] // 0 index is client port
	for {
		timer.Reset(5 * time.Second)
		select {
		case <-timer.C:
			ACKMapClientRecords, _ := ACKMap.Load(clientPort)
			if ACKMapClientRecords != nil {
				finalRecord := ACKMapClientRecords.([]structs.ACKMAPRecord)[len(ACKMapClientRecords.([]structs.ACKMAPRecord))-1]
				if finalRecord.Acked {
					fmt.Println("No timeout for request by ", message.Ports[0])
					return
				}
				fmt.Println("Timeout, resending message")
				helper.SendMessage(message)
			}
		}
	}
}

// Send append request to primary chunk server and wait
func sendChunkAppend(message structs.Message) {
	message.Forward()
	message.Payload = readFile(message.SourceFilename)

	// HTTP Request to Primary Chunk
	fmt.Println("Sending append request to Primary Chunk Server")
	fmt.Println(message.Pointer, message.Ports)
	//success := SendTimerMessage(message)
	helper.SendMessage(message)
}

// Confirm write to the chunk servers
func confirmWrite(message structs.Message) {
	message.Forward()
	message.SetMessageType(helper.DATA_COMMIT)
	helper.SendMessage(message)
}

func finishAppend(message structs.Message) {
	clientPort := message.Ports[0]                    // 0 index is client port
	ACKMapClientRecords, _ := ACKMap.Load(clientPort) // 0 index is client port
	ACKMapClientRecords.([]structs.ACKMAPRecord)[message.RecordIndex].SetAcked(true)
	ACKMap.Store(clientPort, ACKMapClientRecords)

	// Forwarding the ACK_COMMIT to the master
	message.SetPorts([]int{clientPort, helper.MASTER_SERVER_PORT})
	message.Forward()
	helper.SendMessage(message)
}

func InitClient(id int, portNumber int, sourceFilename string, OFSFilename string) {
	fmt.Printf("Client %d is going up at %d\n", id, portNumber)
	go listen(id, portNumber)
	requestMasterAppend(portNumber, sourceFilename, OFSFilename)
	for {

	}
}
