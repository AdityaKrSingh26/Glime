package editor

// defines an opening and closing bracket pair.
type BracketPair struct {
	Open  rune
	Close rune
}

// holds the position of a matching bracket.
type BracketMatch struct {
	Row int
	Col int
}

var bracketPairs = []BracketPair{
	{'(', ')'},
	{'[', ']'},
	{'{', '}'},
}

// finds the matching bracket for the character at (row, col).
// returns nil if the character is not a bracket or no match is found.
func FindMatchingBracket(lines []string, row, col int) *BracketMatch {
	if row < 0 || row >= len(lines) {
		return nil
	}
	line := lines[row]
	if col < 0 || col >= len(line) {
		return nil
	}

	ch := rune(line[col])

	// check if it is an opening bracket
	for _, pair := range bracketPairs {
		if ch == pair.Open {
			return scanForward(lines, row, col, pair.Open, pair.Close)
		}
		if ch == pair.Close {
			return scanBackward(lines, row, col, pair.Close, pair.Open)
		}
	}

	return nil
}

// scans forward from (row, col) to find the matching close bracket.
func scanForward(lines []string, startRow, startCol int, open, close rune) *BracketMatch {
	depth := 0

	for row := startRow; row < len(lines); row++ {
		line := lines[row]
		startC := 0
		if row == startRow {
			startC = startCol
		}

		for col := startC; col < len(line); col++ {
			ch := rune(line[col])
			if ch == open {
				depth++
			} else if ch == close {
				depth--
				if depth == 0 {
					return &BracketMatch{
						Row: row, 
						Col: col,
					}
				}
			}
		}
	}

	return nil
}

// scans backward from (row, col) to find the matching open bracket.
func scanBackward(lines []string, startRow, startCol int, close, open rune) *BracketMatch {
	depth := 0

	for row := startRow; row >= 0; row-- {
		line := lines[row]
		endC := len(line) - 1
		if row == startRow {
			endC = startCol
		}

		for col := endC; col >= 0; col-- {
			ch := rune(line[col])
			if ch == close {
				depth++
			} else if ch == open {
				depth--
				if depth == 0 {
					return &BracketMatch{
						Row: row, 
						Col: col,
					}
				}
			}
		}
	}

	return nil
}
