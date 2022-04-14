package structs

type SuccessfulWrite struct {
	start   int64
	end     int64
}

func GenerateSW(start int64, end int64) SuccessfulWrite {
	SW := SuccessfulWrite{
		start: start,
		end:   end,
	}
	return SW
}
