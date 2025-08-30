package grid

// https://gist.github.com/badboy/6267743
func hash64(ix int64) int32 {
	var x = uint64(ix)
	x = (^x) + (x << 18)
	x = x ^ (x >> 31)
	x = x * 21
	x = x ^ (x >> 11)
	x = x + (x << 6)
	x = x ^ (x >> 22)
	return int32(x)
}
