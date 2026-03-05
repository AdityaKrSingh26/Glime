package syntax

import (
	"fmt"
	"sort"
	"strings"

	"github.com/AdityaKrSingh26/Glime/pkg/ansi"
)

// represents the type of a syntax token.
type TokenType int

const (
	TokenNone TokenType = iota
	TokenKeyword
	TokenString
	TokenComment
	TokenNumber
	TokenFunction
	TokenTypeName
	TokenOperator
	TokenBuiltin
	TokenIdentifier
)

// represents a syntax token with its type and position.
type Token struct {
	Type  TokenType
	Start int // Start position in line (inclusive)
	End   int // End position in line (exclusive)
}

// holds the color indices for syntax highlighting.
type ColorTheme struct {
	Keyword  int
	String   int
	Comment  int
	Number   int
	Function int
	Type     int
	Operator int
	Builtin  int
}

// applies syntax highlighting to source code.
type Highlighter struct {
	lang  *Language
	theme ColorTheme
}

// creates a new highlighter for the given language and theme.
func NewHighlighter(lang *Language, theme ColorTheme) *Highlighter {
	if lang == nil {
		return nil
	}
	return &Highlighter{
		lang:  lang,
		theme: theme,
	}
}

// applies syntax highlighting to a line and returns the colored string.
func (h *Highlighter) Highlight(line string) string {
	if h == nil || h.lang == nil {
		return line
	}

	tokens := h.tokenizeLine(line)
	if len(tokens) == 0 {
		return line
	}

	var result strings.Builder
	pos := 0

	for _, token := range tokens {
		// Write text before token (unhighlighted)
		if token.Start > pos {
			result.WriteString(line[pos:token.Start])
		}

		// Write token with color
		tokenText := line[token.Start:token.End]
		coloredToken := h.colorizeToken(token.Type, tokenText)
		result.WriteString(coloredToken)

		pos = token.End
	}

	// Write remaining text (unhighlighted)
	if pos < len(line) {
		result.WriteString(line[pos:])
	}

	return result.String()
}

// tokenizes a single line of source code.
func (h *Highlighter) tokenizeLine(line string) []Token {
	var tokens []Token

	for _, rule := range h.lang.Rules {
		matches := rule.Pattern.FindAllStringIndex(line, -1)
		for _, match := range matches {
			tokens = append(tokens, Token{
				Type:  rule.TokenType,
				Start: match[0],
				End:   match[1],
			})
		}
	}

	// Sort by start position, prefer longer matches at same position
	sort.Slice(tokens, func(i, j int) bool {
		if tokens[i].Start != tokens[j].Start {
			return tokens[i].Start < tokens[j].Start
		}
		return (tokens[i].End - tokens[i].Start) > (tokens[j].End - tokens[j].Start)
	})

	return h.removeOverlaps(tokens)
}

// removes overlapping tokens, keeping the first/longest at each position.
func (h *Highlighter) removeOverlaps(tokens []Token) []Token {
	if len(tokens) == 0 {
		return tokens
	}

	var result []Token
	lastEnd := 0

	for _, token := range tokens {
		if token.Start < lastEnd {
			continue
		}
		result = append(result, token)
		lastEnd = token.End
	}

	return result
}

// wraps a token with ANSI color codes based on its type.
func (h *Highlighter) colorizeToken(tokenType TokenType, text string) string {
	var color string

	switch tokenType {
	case TokenKeyword:
		color = fmt.Sprintf("%s%s", ansi.SetFgColor(h.theme.Keyword), ansi.Bold)
	case TokenString:
		color = ansi.SetFgColor(h.theme.String)
	case TokenComment:
		color = fmt.Sprintf("%s%s", ansi.SetFgColor(h.theme.Comment), ansi.Italic)
	case TokenNumber:
		color = ansi.SetFgColor(h.theme.Number)
	case TokenFunction:
		color = ansi.SetFgColor(h.theme.Function)
	case TokenTypeName:
		color = ansi.SetFgColor(h.theme.Type)
	case TokenOperator:
		color = ansi.SetFgColor(h.theme.Operator)
	case TokenBuiltin:
		color = ansi.SetFgColor(h.theme.Builtin)
	default:
		return text
	}

	return color + text + ansi.ResetFormat
}
