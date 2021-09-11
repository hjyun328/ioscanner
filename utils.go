package linescanner

func minInt(x int, y int) int {
	if x < y {
		return x
	}
	return y
}

func maxInt(x int, y int) int {
	if x > y {
		return x
	}
	return y
}

func removeCarrageReturn(line []byte) string {
	if len(line) > 0 && line[len(line)-1] == '\r' {
		return string(line[:len(line)-1])
	}
	return string(line)
}
