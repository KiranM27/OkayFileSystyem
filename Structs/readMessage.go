package structs

type ReadMsg struct {
	MessageType      string
	ChunkId          string
	Sources          []int
	SuccessfulWrites []SuccessfulWrite
	Payload          string
}

func GenerateReadMsg(messageType string, chunkId string, sources []int, successfulWrites []SuccessfulWrite, payload string) ReadMsg {
	readMsg := ReadMsg{
		MessageType:      messageType,
		ChunkId:          chunkId,
		Sources:          sources,
		SuccessfulWrites: successfulWrites,
		Payload:          payload,
	}
	return readMsg
}

func (r *ReadMsg) SetMessageType(messageType string) {
	r.MessageType = messageType
}

func (r *ReadMsg) SetChunkId(chunkId string) {
	r.ChunkId = chunkId
}

func (r *ReadMsg) SetSources(sources []int) {
	r.Sources = sources
}

func (r *ReadMsg) SetSuccessfulWrites(successfulWrites []SuccessfulWrite) {
	r.SuccessfulWrites = successfulWrites
}

func (r *ReadMsg) SetPayload(payload string) {
	r.Payload = payload
}