package sensor

const (
	TypeFake = iota + 1
	TypeAtmospheric
)

var sensorByType = map[int]string{
	TypeFake:        "fake",
	TypeAtmospheric: "atmospheric",
}

type Sensor interface {
	// Type denotes the sensor type in the message.
	// If type is 0, it corresponds to a non-data message (ping, etc)
	Type() uint8
	TypeStr() string
	// Init starts or initializes the connection with the sensor
	Init() error
	Ping() error
	// Data returns error and data channels from the sensor.
	// The channels do not have to be buffered
	Data() (chan Data, chan error)
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
