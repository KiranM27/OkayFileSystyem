package structs

type ACKMAPRecord struct {
	RecordIndex int
	ChunkOffset int
	Acked       bool
}

func (r *ACKMAPRecord) SetRecordIndex(recordIndex int) {
	r.RecordIndex = recordIndex
}

func (r *ACKMAPRecord) SetChunkOffset(chunkOffset int) {
	r.ChunkOffset = chunkOffset
}

func (r *ACKMAPRecord) SetAcked(acked bool) {
	r.Acked = acked
}

