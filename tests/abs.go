package main

// ##(--overflow)

// @ ensures res >= 0 && (res == x || res == -x)
func Abs(x int32) (res int32) {
	if x < 0 {
		return -x
	} else {
		return x
	}
}
