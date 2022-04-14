package structs

type SuccessfulWrite struct {
	Start   int64
	End     int64
}

func GenerateSW(start int64, end int64) SuccessfulWrite {
	SW := SuccessfulWrite{
		Start: start,
		End:   end,
	}
	return SW
}
