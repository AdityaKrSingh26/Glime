package editor

// indicates whether the register holds lines or characters.
// 0 - characters
// 1 - lines
type RegisterType int

const (
	RegisterChar RegisterType = iota
	RegisterLine
)

// holds the content of the most recent yank/delete.
type Register struct {
	Content string
	Type    RegisterType
}
