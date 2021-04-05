package result

type Result interface {
	Type() ResultType
	Encode() ([]byte, error)
	Decode(data []byte) error
}

type ResultType int

const (
	EventType ResultType = iota
	ErrorType
)
