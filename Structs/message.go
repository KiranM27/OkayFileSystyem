package structs

import "strconv"

type Message struct {
	MessageType    string
	Ports          []int // 0: Client, 1: Primary, 2+: Secondary
	Pointer        int   // initially 0, increment by 1 everytime received, decrement when ACK
	SourceFilename string
	Filename       string // filename which client requests to append
	ChunkId        string
	Payload        string
	PayloadSize    int64
	ChunkOffset    int64  // Offset at which the data is to be written
	RecordIndex    uint64 // contains the index of the Timeout Indicator (ACKMap) in Client
}

func (message Message) GenerateUid() string {
	uid := message.Filename + message.ChunkId + strconv.Itoa(int(message.ChunkOffset))
	return uid
}

func (message *Message) Forward() {
	message.Pointer += 1
}

func (message *Message) Reply() {
	message.Pointer -= 1
}

func (message *Message) SetMessageType(messageType string) {
	message.MessageType = messageType
}

func (message *Message) SetPorts(ports []int) {
	message.Ports = ports
}

func (message *Message) SetPayload(payload string) {
	message.Payload = payload
}

func (message *Message) SetChunkOffset(offset int64) {
	message.ChunkOffset = offset
}
