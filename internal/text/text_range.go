package text

type TextRange struct {
	start TextSize
	end   TextSize
}

func NewTextRange(start, end TextSize) TextRange {
	if start > end {
		panic("text: range start must be less than or equal to end")
	}
	return TextRange{
		start: start,
		end:   end,
	}
}

func (r TextRange) Start() TextSize {
	return r.start
}

func (r TextRange) End() TextSize {
	return r.end
}

func (r TextRange) Len() TextSize {
	return r.end - r.start
}

func (r TextRange) IsEmpty() bool {
	return r.start == r.end
}

func (r TextRange) Contains(offset TextSize) bool {
	return r.start <= offset && offset < r.end
}
