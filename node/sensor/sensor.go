package sensor

const (
	TypeFake = iota + 1
)

var sensorByType = map[int]string{
	TypeFake: "fake",
}

type Sensor interface {
	// Type denotes the sensor type in the message.
	// If type is 0, it corresponds to a non-data message (ping, etc)
	Type() uint8
	TypeStr() string
	Ping() error
	// PollRate returns the desired poll rate for the sensor in milliseconds
	PollRate() int64
	Poll() (Data, error)
	Close()
}

// Data is the wire format for sensor readings
type Data struct {
	Typ  uint8
	Data []byte
}

// NameByType maps human readable names to sensor types
func NameByType(typ int) string {
	if n, ok := sensorByType[typ]; ok {
		return n
	} else {
		return "invalid type"
	}
}
