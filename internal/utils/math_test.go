package utils

import (
	"testing"
)

func TestMaxInt(t *testing.T) {
	a := 1
	b := 2
	expected := b

	actual := Max[int](a, b)
	if expected != actual {
		t.Errorf("Incorrect result. Actual: %d, Expected: %d", actual, expected)
	}

	actual = Max[int](b, a)
	if expected != actual {
		t.Errorf("Incorrect result. Actual: %d, Expected: %d", actual, expected)
	}
}

func TestClamp(t *testing.T) {
	max := 2
	min := 0

	value := 1
	expected := value
	actual := Clamp[int](value, min, max)
	if expected != actual {
		t.Errorf("Incorrect result. Actual: %d, Expected: %d", actual, expected)
	}

	value = -1
	expected = min
	actual = Clamp[int](value, min, max)
	if expected != actual {
		t.Errorf("Incorrect result. Actual: %d, Expected: %d", actual, expected)
	}

	value = 3
	expected = max
	actual = Clamp[int](value, min, max)
	if expected != actual {
		t.Errorf("Incorrect result. Actual: %d, Expected: %d", actual, expected)
	}
}
