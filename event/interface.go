package event

type IEvent interface {
	Print() string
	Encode() ([]byte, error)
}
