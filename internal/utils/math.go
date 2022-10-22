package utils

func ClampFloat32(value float32, left float32, right float32) float32 {
	if value > right {
		return right
	} else if value < left {
		return left
	}

	return value
}

func ClampInt(value int, left int, right int) int {
	if value > right {
		return right
	} else if value < left {
		return left
	}

	return value
}

func MaxInt(left int, right int) int {
	if left > right {
		return left
	} else {
		return right
	}
}
