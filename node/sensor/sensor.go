package sensor

const (
	TypeFake = iota
)

type Sensor interface {
	Type() uint8
	Ping() error
	Poll() ([]byte, error)
	Close()
}
