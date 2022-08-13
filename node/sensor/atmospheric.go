package sensor

import "os/exec"

// AtmosphericSensor tracks sensor input from the BME280 and SGP40 sensor boards from Adafruit
type AtmosphericSensor struct {
	sensorProc exec.Cmd
}

func (a *AtmosphericSensor) Type() uint8 {
	return TypeAtmospheric
}

func (a *AtmosphericSensor) TypeStr() string {
	return NameByType(int(a.Type()))
}

func (a *AtmosphericSensor) Ping() error {
	//TODO implement me
	panic("implement me")
}

func (a *AtmosphericSensor) Init() error {
	//TODO implement me
	panic("implement me")
}

func (a *AtmosphericSensor) PollRate() int64 {
	return 1000
}

func (a *AtmosphericSensor) Poll() (Data, error) {
	//TODO implement me
	panic("implement me")
}

func (a *AtmosphericSensor) Close() {
	//TODO implement me
	panic("implement me")
}
