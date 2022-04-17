package client

// Client works on only one append operration at a time !!

import (
	"encoding/json"
	"fmt"
	"net/http"
	helper "oks/Helper"
	structs "oks/Structs"
	"strconv"
	"sync"
	"time"
	"github.com/gin-gonic/gin"
)

var ACKMap sync.Map

func listen(id int, portNumber int) {
	router := gin.New()
    router.Use(
        gin.LoggerWithWriter(gin.DefaultWriter, "/message"),
        gin.Recovery(),
    )
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
	context.IndentedJSON(http.StatusOK, "Message received by the client")

	switch message.MessageType {
	case helper.DATA_APPEND:
		go sendChunkAppend(message)
	case helper.ACK_APPEND:
		go confirmWrite(message)
	case helper.ACK_COMMIT:
		go finishAppend(message)
	}
}

func readChunk(chunkId string) string {
	readMsg := structs.GenerateReadMsgV2(helper.READ_REQ_TO_MASTER, chunkId)
	resBody := helper.SendReadMsg(readMsg, helper.MASTER_SERVER_PORT)
	var recReadMsg structs.ReadMsg
	json.Unmarshal(resBody, &recReadMsg)
	recReadMsg.SetMessageType(helper.READ_REQ_TO_CHUNK)
	content := helper.SendReadMsg(recReadMsg, recReadMsg.Sources[0])
	return string(content)
}

// Send a request to Master that client wants to append
// The message has the size of the data
func requestMasterAppend(clientPort int, sourceFilename string, OFSFilename string) {
	var numChunks uint64
	//var message structs.Message

	// Call helper function to read source file size
	sourceFileByteSize := getFileSize(sourceFilename)

	// Check byte size of file. If it is more than 2.5kb - split the file ito chunks
	if sourceFileByteSize > helper.CHUNK_SIZE {
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
		fmt.Println("Client " + strconv.Itoa(clientPort) + " is sending append request to Master for file " + sourceFilename)
		helper.SendMessage(message)
		go runTimer(message)

	} else {
		sourceFilePrefix := helper.RemoveExtension(sourceFilename)

		for i := uint64(0); i < numChunks; i++ {
			smallFileName := sourceFilePrefix + strconv.FormatUint(i, 10) + ".txt"
			smallFileSize := getFileSize(smallFileName)

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
				SourceFilename: smallFileName,
				Filename:       OFSFilename, // TODO: This should be the same file name still right
				PayloadSize:    smallFileSize,
				RecordIndex:    uint64(recordIndex),
			}
			ACKMap.Store(clientPort, ACKMapClientRecords)
			fmt.Println(strconv.Itoa(clientPort) + "is sending append request to Master for file " + smallFileName)
			helper.SendMessage(message)
			go runTimer(message)
		}

	}
}

func runTimer(message structs.Message) {
	timer := time.NewTimer(helper.DEFAULT_TIMEOUT)
	clientPort := message.Ports[0] // 0 index is client port
	for {
		timer.Reset(helper.DEFAULT_TIMEOUT)
		select {
		case <-timer.C:
			ACKMapClientRecords, _ := ACKMap.Load(clientPort)
			if ACKMapClientRecords != nil {
				//finalRecord := ACKMapClientRecords.([]structs.ACKMAPRecord)[len(ACKMapClientRecords.([]structs.ACKMAPRecord))-1]
				finalRecord := ACKMapClientRecords.([]structs.ACKMAPRecord)[message.RecordIndex]
				if finalRecord.Acked {
					fmt.Println("Client ", clientPort, ": No timeout for ", message.MessageType,  " request of file - ", message.Filename)
					return
				}
				fmt.Println("Client ", clientPort, ": Timeout for ", message.MessageType,  " request of file - ", message.Filename, ". Resending request now.")
				helper.SendMessage(message)
			}
		}
	}
}

// Send append request to primary chunk server and wait
func sendChunkAppend(message structs.Message) {
	message.Forward()
	message.Payload = readFile(message.SourceFilename)
	clientPort := message.Ports[0]
	PCS := message.Ports[1]
	fileName := message.Filename

	// HTTP Request to Primary Chunk
	fmt.Println("Client ", clientPort, ": Sending append request to Primary Chunk Server - ", PCS, " - for file - ", fileName)
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

func InitWriteClient(id int, portNumber int, sourceFilename string, OFSFilename string) {
	go listen(id, portNumber)
	// start := time.Now()
	requestMasterAppend(portNumber, sourceFilename, OFSFilename)
	// end := time.Now()
	// fmt.Printf("Start Time for client %d : %.2f\n", id, start)
	// fmt.Printf("End Time for client %d : %.2f\n", id, end)
	// fmt.Printf("Append Time taken = %.2f seconds \n", end.Sub(start).Seconds())
	for {

	}
}

func InitReadClient(id int, portNumber int, chunkId string) {
	go listen(id, portNumber)
	start := time.Now()
	content := readChunk(chunkId)
	fmt.Println("The follwing is the text that was read from chunk with chunkId - ", chunkId, " - Data - ", content)
	end := time.Now()
	fmt.Printf("Start Time for client %d : %.2f\n", id, start)
	fmt.Printf("End Time for client %d : %.2f\n", id, end)

	fmt.Printf("Read Time taken = %.2f seconds \n", end.Sub(start).Seconds())
	for {

	}
}
