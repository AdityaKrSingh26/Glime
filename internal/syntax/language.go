package syntax

import (
	"path/filepath"
	"regexp"
	"strings"
)

// Rule represents a syntax highlighting rule with a regex pattern and token type.
type Rule struct {
	Pattern   *regexp.Regexp
	TokenType TokenType
}

// Language defines syntax rules for a programming language.
type Language struct {
	Name  string
	Rules []Rule
}

// LanguageName returns the display name for a file based on its extension.
// Returns "" if the file has no recognized extension.
func LanguageName(filename string) string {
	if filename == "" {
		return ""
	}

	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".go":
		return "Go"
	case ".js":
		return "JavaScript"
	case ".ts":
		return "TypeScript"
	case ".py":
		return "Python"
	case ".rs":
		return "Rust"
	case ".java":
		return "Java"
	case ".c":
		return "C"
	case ".cpp", ".cc", ".cxx":
		return "C++"
	case ".h", ".hpp":
		return "Header"
	case ".md":
		return "Markdown"
	case ".json":
		return "JSON"
	case ".yaml", ".yml":
		return "YAML"
	case ".toml":
		return "TOML"
	case ".html":
		return "HTML"
	case ".css":
		return "CSS"
	case ".sh", ".bash":
		return "Shell"
	default:
		return ""
	}
}

// DetectLanguage detects the language based on file extension.
func DetectLanguage(filename string) *Language {
	if filename == "" {
		return nil
	}

	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".go":
		return GoLanguage()
	case ".js":
		return JavaScriptLanguage()
	case ".py":
		return PythonLanguage()
	default:
		return nil
	}
}

// GoLanguage returns the syntax rules for Go.
func GoLanguage() *Language {
	rules := []Rule{
		// Comments (must be before operators to handle //)
		{regexp.MustCompile(`//.*$`), TokenComment},
		{regexp.MustCompile(`/\*.*?\*/`), TokenComment},

		// Strings
		{regexp.MustCompile("`[^`]*`"), TokenString},           // Raw strings
		{regexp.MustCompile(`"(?:[^"\\]|\\.)*"`), TokenString}, // Quoted strings
		{regexp.MustCompile(`'(?:[^'\\]|\\.)*'`), TokenString}, // Char literals

		// Numbers
		{regexp.MustCompile(`\b0[xX][0-9a-fA-F]+\b`), TokenNumber}, // Hex
		{regexp.MustCompile(`\b\d+\.?\d*([eE][+-]?\d+)?\b`), TokenNumber},

		// Keywords
		{regexp.MustCompile(`\b(?:break|case|chan|const|continue|default|defer|else|fallthrough|for|func|go|goto|if|import|interface|map|package|range|return|select|struct|switch|type|var)\b`), TokenKeyword},

		// Builtins
		{regexp.MustCompile(`\b(?:append|cap|close|complex|copy|delete|imag|len|make|new|panic|print|println|real|recover|bool|byte|complex64|complex128|error|float32|float64|int|int8|int16|int32|int64|rune|string|uint|uint8|uint16|uint32|uint64|uintptr|true|false|nil|iota)\b`), TokenBuiltin},

		// Types (capitalized identifiers often indicate types in Go)
		{regexp.MustCompile(`\b[A-Z][a-zA-Z0-9_]*\b`), TokenTypeName},

		// Function calls
		{regexp.MustCompile(`\b[a-z_][a-zA-Z0-9_]*\s*(?=\()`), TokenFunction},

		// Operators
		{regexp.MustCompile(`[+\-*/%&|^<>=!:]+`), TokenOperator},
	}

	return &Language{
		Name:  "Go",
		Rules: rules,
	}
}

// JavaScriptLanguage returns the syntax rules for JavaScript.
func JavaScriptLanguage() *Language {
	rules := []Rule{
		// Comments
		{regexp.MustCompile(`//.*$`), TokenComment},
		{regexp.MustCompile(`/\*.*?\*/`), TokenComment},

		// Strings
		{regexp.MustCompile("`[^`]*`"), TokenString},           // Template strings
		{regexp.MustCompile(`"(?:[^"\\]|\\.)*"`), TokenString}, // Double quoted
		{regexp.MustCompile(`'(?:[^'\\]|\\.)*'`), TokenString}, // Single quoted

		// Numbers
		{regexp.MustCompile(`\b0[xX][0-9a-fA-F]+\b`), TokenNumber},
		{regexp.MustCompile(`\b\d+\.?\d*([eE][+-]?\d+)?\b`), TokenNumber},

		// Keywords
		{regexp.MustCompile(`\b(?:async|await|break|case|catch|class|const|continue|debugger|default|delete|do|else|export|extends|finally|for|function|if|import|in|instanceof|let|new|return|static|super|switch|this|throw|try|typeof|var|void|while|with|yield)\b`), TokenKeyword},

		// Builtins
		{regexp.MustCompile(`\b(?:Array|Boolean|Date|Error|Function|JSON|Math|Number|Object|Promise|RegExp|String|Symbol|console|document|window|null|undefined|true|false|NaN|Infinity)\b`), TokenBuiltin},

		// Function calls
		{regexp.MustCompile(`\b[a-zA-Z_$][a-zA-Z0-9_$]*\s*(?=\()`), TokenFunction},

		// Operators
		{regexp.MustCompile(`[+\-*/%&|^<>=!:?]+`), TokenOperator},
	}

	return &Language{
		Name:  "JavaScript",
		Rules: rules,
	}
}

// PythonLanguage returns the syntax rules for Python.
func PythonLanguage() *Language {
	rules := []Rule{
		// Comments
		{regexp.MustCompile(`#.*$`), TokenComment},

		// Strings (triple quotes first to match longer patterns)
		{regexp.MustCompile(`"""[^"]*"""`), TokenString},
		{regexp.MustCompile(`'''[^']*'''`), TokenString},
		{regexp.MustCompile(`"(?:[^"\\]|\\.)*"`), TokenString},
		{regexp.MustCompile(`'(?:[^'\\]|\\.)*'`), TokenString},

		// Numbers
		{regexp.MustCompile(`\b0[xX][0-9a-fA-F]+\b`), TokenNumber},
		{regexp.MustCompile(`\b\d+\.?\d*([eE][+-]?\d+)?\b`), TokenNumber},

		// Keywords
		{regexp.MustCompile(`\b(?:and|as|assert|break|class|continue|def|del|elif|else|except|finally|for|from|global|if|import|in|is|lambda|nonlocal|not|or|pass|raise|return|try|while|with|yield)\b`), TokenKeyword},

		// Builtins
		{regexp.MustCompile(`\b(?:abs|all|any|bin|bool|bytes|chr|dict|dir|enumerate|filter|float|hex|int|isinstance|len|list|map|max|min|open|print|range|reversed|set|sorted|str|sum|tuple|type|zip|True|False|None)\b`), TokenBuiltin},

		// Function definitions
		{regexp.MustCompile(`\bdef\s+([a-zA-Z_][a-zA-Z0-9_]*)`), TokenFunction},

		// Function calls
		{regexp.MustCompile(`\b[a-zA-Z_][a-zA-Z0-9_]*\s*(?=\()`), TokenFunction},

		// Operators
		{regexp.MustCompile(`[+\-*/%&|^<>=!:]+`), TokenOperator},
	}

	return &Language{
		Name:  "Python",
		Rules: rules,
	}
}
