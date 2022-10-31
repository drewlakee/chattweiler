package utils

func Clamp[T float32 | int](value, min, max T) T {
	if value > max {
		return max
	} else if value < min {
		return min
	}

	return value
}

func Max[T int](left, right T) T {
	if left > right {
		return left
	} else {
		return right
	}
}
