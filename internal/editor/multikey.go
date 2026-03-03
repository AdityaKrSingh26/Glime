package editor

// PendingCommand tracks multi-key command state (e.g., count + operator + motion).
type PendingCommand struct {
	Count    int  // numeric prefix
	Operator rune // pending operator
	HasCount bool // whether a count has been started
}

func (p *PendingCommand) Reset() {
	p.Count = 0
	p.Operator = 0
	p.HasCount = false
}

// returns the count, default to 1 if none specified.
func (p *PendingCommand) EffectiveCount() int {
	if p.Count == 0 {
		return 1
	}
	return p.Count
}

// adds a digit to the count prefix.
func (p *PendingCommand) AccumulateDigit(d int) {
	p.Count = p.Count*10 + d
	p.HasCount = true
}
