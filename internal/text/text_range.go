package text

// TextRange represents a range of text with a start and end offset.
type TextRange struct {
	start TextSize
	end   TextSize
}

// NewTextRange creates a new TextRange with the given start and end offsets.
func NewTextRange(start, end TextSize) TextRange {
	if start > end {
		panic("text: range start must be less than or equal to end")
	}
	return TextRange{
		start: start,
		end:   end,
	}
}

// Start returns the start offset of the text range.
func (r TextRange) Start() TextSize {
	return r.start
}

// End returns the end offset of the text range.
func (r TextRange) End() TextSize {
	return r.end
}

// Len returns the length of the text range.
func (r TextRange) Len() TextSize {
	return r.end - r.start
}

// IsEmpty returns true if the text range is empty.
func (r TextRange) IsEmpty() bool {
	return r.start == r.end
}

// Contains returns true if the given offset is within the text range.
func (r TextRange) Contains(offset TextSize) bool {
	return r.start <= offset && offset < r.end
}
