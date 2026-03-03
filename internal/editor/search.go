package editor

import "strings"

// indicates forward or backward search.
// 0 - forward
// 1 - backward
type SearchDirection int

const (
	SearchForward SearchDirection = iota
	SearchBackward
)

// represents a single match location.
type SearchMatch struct {
	Row      int
	ColStart int
	ColEnd   int
}

// holds the current search pattern and results.
type SearchState struct {
	Pattern      string
	Direction    SearchDirection
	Matches      []SearchMatch
	CurrentIndex int
	Active       bool
}

// finds all occurrences of the pattern in the given lines.
func (s *SearchState) FindAll(lines []string) {
	s.Matches = s.Matches[:0]

	if s.Pattern == "" {
		return
	}

	for row, line := range lines {
		offset := 0
		for {
			idx := strings.Index(line[offset:], s.Pattern)
			if idx < 0 {
				break
			}
			start := offset + idx
			end := start + len(s.Pattern)
			s.Matches = append(s.Matches, SearchMatch{
				Row:      row,
				ColStart: start,
				ColEnd:   end,
			})
			offset = end
		}
	}
}

// finds the next match after (row, col) in the given direction, wraps around the file
func (s *SearchState) NextMatch(row, col int) int {
	if len(s.Matches) == 0 {
		return -1
	}

	if s.Direction == SearchForward {
		// first match after current position
		for i, m := range s.Matches {
			if m.Row > row || (m.Row == row && m.ColStart > col) {
				return i
			}
		}
		return 0 // wrap to first match
	}

	// find last match before current position
	for i := len(s.Matches) - 1; i >= 0; i-- {
		m := s.Matches[i]
		if m.Row < row || (m.Row == row && m.ColStart < col) {
			return i
		}
	}
	return len(s.Matches) - 1 // wrap to last match
}

// finds the previous match.
func (s *SearchState) PrevMatch(row, col int) int {
	if len(s.Matches) == 0 {
		return -1
	}

	if s.Direction == SearchForward {
		// find last match before current position
		for i := len(s.Matches) - 1; i >= 0; i-- {
			m := s.Matches[i]
			if m.Row < row || (m.Row == row && m.ColStart < col) {
				return i
			}
		}
		return len(s.Matches) - 1
	}

	// find first match after current position
	for i, m := range s.Matches {
		if m.Row > row || (m.Row == row && m.ColStart > col) {
			return i
		}
	}
	return 0
}

// returns the matches on a specific row.
func (s *SearchState) MatchesForRow(row int) []SearchMatch {
	var result []SearchMatch
	for _, m := range s.Matches {
		if m.Row == row {
			result = append(result, m)
		}
	}
	return result
}
