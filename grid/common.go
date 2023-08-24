package grid

func nextPow2(x uint32) uint32 {
	x--
	x |= x >> 1
	x |= x >> 2
	x |= x >> 4
	x |= x >> 8
	x |= x >> 16
	return x + 1
}

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
func hash32(ix int32) int32 {
	var x = uint32(ix)
	x += ^(x << 15)
	x ^= x >> 10
	x += x << 3
	x ^= x >> 6
	x += ^(x << 11)
	x ^= x >> 16
	return int32(x)
}

func memset(a []int32, x int32) {
	if len(a) == 0 {
		return
	}
	a[0] = x
	for bp := 1; bp < len(a); bp <<= 1 {
		copy(a[bp:], a[:bp])
	}
}
