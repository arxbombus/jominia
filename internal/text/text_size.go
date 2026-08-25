package text

import "math"

type TextSize uint32

func SizeOf(value string) TextSize {
	size := len(value)

	if int64(size) > math.MaxUint32 {
		panic("text: value exceeds maximum TextSize")
	}

	return TextSize(size)
}
