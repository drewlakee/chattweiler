package math

func ClampFloat32(value float32, left float32, right float32) float32 {
	if value > right {
		return right
	} else if value < left {
		return left
	}

	return value
}
