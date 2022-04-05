package structs

type RepMsg struct {
	Index       int
	MessageType string
	ChunkId     string
	Sources     []int
	TargetCS    int
	Payload 	string
}

func (r *RepMsg) SetPayload(payload string) {
	r.Payload = payload
}

func (r *RepMsg) SetMessageType(messageType string) {
	r.MessageType = messageType
}



