package text

import "math"

// TextSize represents the length of text.
type TextSize uint32

// SizeOf returns the TextSize corresponding to the length of the given string. It panics if the string's length exceeds the maximum value (math.MaxUint32) representable by TextSize.
func SizeOf(value string) TextSize {
	size := len(value)
	if int64(size) > math.MaxUint32 {
		panic("text: value exceeds maximum TextSize")
	}
	return TextSize(size)
}
