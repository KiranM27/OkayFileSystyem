package structs

type RepMsg struct {
	Index       int
	MessageType string
	ChunkId     string
	Sources     []int
	TargetCS    int
	Payload 	string
}

