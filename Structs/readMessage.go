package structs

type ReadMessage struct {
	Index       int
	MessageType string
	ChunkId     string
	Sources     []int
	TargetCS    int
	Payload 	string
}
