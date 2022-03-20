package main

import (
	"fmt"
	"net/http"
	"github.com/gin-gonic/gin"
	//client "oks/Client"
	helper "oks/Helper"
	structs "oks/Structs"
	chunkServer "oks/ChunkServer"
	"strconv"
	)

var metaData MetaData
var port_map Port

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
}

type Port struct{
	portToInt map[string]int
}


func listen(nodePid int, portNo int) {
	router := gin.Default()
	router.POST("/message", postMessageHandler)

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
	context.IndentedJSON(http.StatusOK, message.MessageType+" message from Node "+strconv.Itoa(message.Ports[message.Pointer])+" was received by Master")

	switch message.MessageType{
		case helper.DATA_APPEND:
			go appendMessageHandler(message)
	}
}

func appendMessageHandler(message structs.Message){
	fmt.Println("Before Sending")
	message1 := structs.Message{
		MessageType: helper.DATA_APPEND,
		Ports: []int{port_map.portToInt["6"], port_map.portToInt["1"], port_map.portToInt["2"], port_map.portToInt["3"]},
		Pointer: 0,
		Filename: message.Filename,
		ChunkId: metaData.fileIdToChunkId[message.Filename][0],
		Payload: message.Payload,
		PayloadSize: message.PayloadSize,
		ChunkOffset: metaData.chunkIdToOffset[metaData.fileIdToChunkId[message.Filename][0]],
	}
	fmt.Println("Master replying append request to client")
	helper.SendMessage(message1)

	// increment offset
	new_offset := message1.ChunkOffset + message1.PayloadSize
	metaData.chunkIdToOffset[metaData.fileIdToChunkId[message.Filename][0]] = new_offset

}



func main(){
	metaData.fileIdToChunkId = make(map[string][]string)
	metaData.chunkIdToChunkserver = make(map[string][]int)
	metaData.chunkIdToOffset = make(map[string]int64)
	port_map.portToInt = map[string]int{"0":  8080, "1": 8081, "2": 8082, "3": 8083, "4": 8084, "5": 8085, "6": 8086}

	// create dummy data
	metaData.fileIdToChunkId["test.txt"] = []string{"test_c0"}
	metaData.chunkIdToChunkserver["test_c0"] = []int{8081, 8082, 8083}
	offset := int64(0)
	metaData.chunkIdToOffset["test_c0"] = offset

	go chunkServer.ChunkServer(1, 8081)
	go chunkServer.ChunkServer(2, 8082)
	go chunkServer.ChunkServer(3, 8083)
	listen(0, 8080)
	
}