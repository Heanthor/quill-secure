package sensor

import (
	"github.com/rs/zerolog/log"
	"io"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

type AtmosphericDataLine struct {
	Timetsamp   time.Time
	Temperature float32
	Humidity    float32
	Pressure    float32
	Altitude    float32
	VOCIndex    float32
}

// AtmosphericSensor tracks sensor input from the BME280 and SGP40 sensor boards from Adafruit
type AtmosphericSensor struct {
	sensorProc     *exec.Cmd
	sensorStdout   io.Reader
	executablePath string
}

func NewAtmospheric(executable string) *AtmosphericSensor {
	return &AtmosphericSensor{
		executablePath: executable,
	}
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
	args := strings.Split(a.executablePath, " ")
	log.Debug().Str("path", a.executablePath).Msg("Atmospheric executable path")
	cmd := exec.Command(args[0], args[1:]...)
	a.sensorProc = cmd

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	a.sensorStdout = stdout

	// start nonblocking
	if err := cmd.Start(); err != nil {
		return err
	}

	return nil
}

func (a *AtmosphericSensor) Data() (chan Data, chan error) {
	errCh := make(chan error)
	dataCh := make(chan Data)

	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := a.sensorStdout.Read(buf)
			if err != nil {
				if err == io.EOF {
					// sensor program died or program exited
				} else {
					errCh <- err
				}
			}

			if n == 0 {
				continue
			}

			sensorLine := buf[:n]
			lm := log.Debug().Str("line", string(sensorLine))
			if !validateSensorLine(string(sensorLine)) {
				lm.Msg("invalid sensor line")
				continue
			}

			d := Data{
				Typ:  a.Type(),
				Data: sensorLine,
			}
			dataCh <- d
			lm.Msg("sensor line")
			time.Sleep(100 * time.Millisecond)
		}
	}()

	return dataCh, errCh
}

func validateSensorLine(line string) bool {
	// 1660369274.4389172,25.2296875,43.159619678029735,1009.1293692371094,70.40053920398444,0
	return strings.Count(line, ",") == 5
}

func (a *AtmosphericSensor) Close() {
	// TODO kill the proc
	log.Info().Msg("close atmospheric sensor process")
	a.sensorProc.Process.Signal(syscall.SIGINT)
	a.sensorProc.Wait()
}
