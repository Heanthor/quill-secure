package sensor

const (
	TypeFake = iota + 1
)

type Sensor interface {
	// Type denotes the sensor type in the message.
	// If type is 0, it corresponds to a non-data message (ping, etc)
	Type() uint8
	Ping() error
	// PollRate returns the desired poll rate for the sensor in milliseconds
	PollRate() int64
	Poll() ([]byte, error)
	Close()
}
