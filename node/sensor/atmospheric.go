package sensor

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// AtmosphericSensor tracks sensor input from the BME280 and SGP40 sensor boards from Adafruit
type AtmosphericSensor struct {
	sensorProc     *exec.Cmd
	sensorStdout   io.Reader
	executablePath string
	pollFreq       int
}

type AtmosphericDataLine struct {
	Timestamp   time.Time
	Temperature float32
	Humidity    float32
	Pressure    float32
	Altitude    float32
	VOCIndex    float32
}

func NewAtmospheric(executable string, sensorPollFreqSec int) *AtmosphericSensor {
	return &AtmosphericSensor{
		executablePath: executable,
		pollFreq:       sensorPollFreqSec,
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
	args = append(args, "--poll-frequency="+strconv.Itoa(a.pollFreq))
	fmt.Println(args)
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
				time.Sleep(100 * time.Millisecond)
				continue
			}

			sensorLine := string(buf[:n])
			sensorLine = strings.TrimSpace(sensorLine)
			lm := log.Debug().Str("line", sensorLine)
			if !validateSensorLine(sensorLine) {
				lm.Msg("invalid sensor line")
				continue
			}

			d := Data{
				Typ:  a.Type(),
				Data: []byte(sensorLine),
			}
			dataCh <- d
			lm.Msg("sensor line")
			time.Sleep(100 * time.Millisecond)
		}
	}()

	return dataCh, errCh
}

func validateSensorLine(line string) bool {
	// 1660369274,25.2296875,43.159619678029735,1009.1293692371094,70.40053920398444,0
	return strings.Count(line, ",") == 5
}

// ParseValidAtmosphericSensorLine assumes the sensor line is well-formed.
// If a part is missing, an empty struct is returned.
// If all parts are present but not valid types, that field will be unset.
func ParseValidAtmosphericSensorLine(line string) AtmosphericDataLine {
	parts := strings.Split(line, ",")
	if len(parts) != 6 {
		return AtmosphericDataLine{}
	}

	tsInt, err := strconv.Atoi(parts[0])
	if err != nil {
		tsInt = 0
	}

	return AtmosphericDataLine{
		Timestamp:   time.Unix(int64(tsInt), 0),
		Temperature: parseFloat32(parts[1]),
		Humidity:    parseFloat32(parts[2]),
		Pressure:    parseFloat32(parts[3]),
		Altitude:    parseFloat32(parts[4]),
		VOCIndex:    parseFloat32(parts[5]),
	}
}

func parseFloat32(part string) float32 {
	f, err := strconv.ParseFloat(part, 32)
	if err != nil {
		f = 0
	}

	return float32(f)
}

func (a *AtmosphericSensor) Close() {
	// TODO kill the proc
	log.Info().Msg("close atmospheric sensor process")
	a.sensorProc.Process.Signal(syscall.SIGINT)
	a.sensorProc.Wait()
}
