package main

import (
	"fmt"
	mynet "github.com/Heanthor/quill-secure/net"
	"github.com/Heanthor/quill-secure/node/sensor"
	"github.com/rs/zerolog/log"
	"io"
	"net"
	"time"
)

type SensorCollection struct {
	deviceID      uint8
	activeSensors []sensor.Sensor
	sensorPings   chan sensorDataWrapper
	errorPings    chan sensorErrorWrapper

	leader        mynet.Dest
	pingTicker    *time.Ticker
	leaderHealthy bool

	sendChan chan outgoingPacketWrapper
	doneChan chan bool
}

type outgoingPacketWrapper struct {
	p mynet.Packet
	s sensor.Sensor
}

// NewSensorCollection creates resources, but does not start any polling or processing
func NewSensorCollection(deviceID uint8, host string, port, pingIntervalSecs, packetBufferSize int) SensorCollection {
	t := time.NewTicker(time.Duration(pingIntervalSecs) * time.Second)
	sendChan := make(chan outgoingPacketWrapper, packetBufferSize)
	ip, err := mynet.ParseHost(host)
	if err != nil {
		log.Fatal().Msg("Invalid host parameter")
	}

	return SensorCollection{
		deviceID: deviceID,
		leader: mynet.Dest{
			Host: ip,
			Port: port,
		},
		sensorPings: make(chan sensorDataWrapper),
		errorPings:  make(chan sensorErrorWrapper),
		doneChan:    make(chan bool),
		sendChan:    sendChan,
		pingTicker:  t,
	}
}

type sensorDataWrapper struct {
	sensor sensor.Sensor
	data   sensor.Data
}

type sensorErrorWrapper struct {
	sensor sensor.Sensor
	err    error
}

func (s *SensorCollection) RegisterSensors(atmosphericExecutable string, atmosphericPollFreq int) {
	//s.registerSensor(&sensor.FakeSensor{Buf: "fake data"})
	s.registerSensor(sensor.NewAtmospheric(atmosphericExecutable, atmosphericPollFreq))
}

func (s *SensorCollection) registerSensor(sn sensor.Sensor) {
	if err := sn.Init(); err != nil {
		log.Fatal().Err(err).Str("type", sn.TypeStr()).Msg("Sensor failed to initialize")
		return
	}

	log.Info().Str("type", sn.TypeStr()).Msg("Registered new sensor")

	// consolidate pings from sensors into single channels
	go func(sn sensor.Sensor) {
		data, err := sn.Data()
		for {
			select {
			case <-s.doneChan:
				return
			case d := <-data:
				s.sensorPings <- sensorDataWrapper{
					sensor: sn,
					data:   d,
				}
			case e := <-err:
				s.errorPings <- sensorErrorWrapper{
					sensor: sn,
					err:    e,
				}
			}
		}
	}(sn)

	s.activeSensors = append(s.activeSensors, sn)
}

func (s *SensorCollection) StopPolling() {
	for _, sn := range s.activeSensors {
		sn.Close()
	}

	close(s.sensorPings)
	close(s.sendChan)
}

// StartLeaderHealthCheck pings the leader node every pingIntervalSecs seconds
func (s *SensorCollection) StartLeaderHealthCheck() {
	go func() {
		for {
			select {
			case <-s.pingTicker.C:
				if err := s.pingLeader(); err != nil {
					// TODO maybe use a mutex here
					s.leaderHealthy = false
					log.Err(err).Msg("Cannot reach leader node")
				} else {
					if s.leaderHealthy == false {
						log.Info().Msg("Leader node reachable again")
					}
					s.leaderHealthy = true
				}
			}
		}
	}()
}

// pingLeader opens a TCP conn, sends a small amount of data, and returns error if the message is not acked.
func (s *SensorCollection) pingLeader() error {
	p := mynet.Packet{
		UID: s.deviceID,
		Typ: mynet.PacketTypeAnnounce,
	}

	return s.SendPacket(p)
}

// sendConsumer reads from sendChan and tries to send packets, if the leader is healthy.
// If leader is not healthy, packets remain buffered in the channel
func (s *SensorCollection) sendConsumer() {
	for {
		if s.leaderHealthy {
			select {
			case w, more := <-s.sendChan:
				if !more {
					s.doneChan <- true
					return
				} else {
					if err := s.SendPacket(w.p); err != nil {
						// TODO maybe use an error chan
						log.Err(err).Uint8("sensor", w.s.Type()).Msg("Error sending packet")
					}
				}
			}
		}
	}
}

// Poll polls connected sensors, and sends any new data to the leader node
// This method blocks on sensorPings
func (s *SensorCollection) Poll() {
	go s.sendConsumer()

	for sn := range s.sensorPings {
		data := sn.data
		// send data to leader
		p := mynet.Packet{
			UID:  s.deviceID,
			Typ:  mynet.PacketTypeSensorData,
			Data: data,
		}

		select {
		case s.sendChan <- outgoingPacketWrapper{p, sn.sensor}:
		default:
			log.Warn().Msg("Sensor buffer is full, dropping packet")
		}
	}
}

// SendPacket opens a TCP conn to leader, sends the packet, and closes the connection
func (s *SensorCollection) SendPacket(p mynet.Packet) error {
	log.Debug().Interface("packet", p).Msg("SendPacket")
	conn, err := net.DialTCP("tcp", nil, &net.TCPAddr{
		IP:   s.leader.Host,
		Port: s.leader.Port,
	})
	if err != nil {
		return fmt.Errorf("error creating TCP conn to leader: %w", err)
	}
	defer conn.Close()

	if err := sendPacket(p, conn); err != nil {
		return fmt.Errorf("error sending packet: %w", err)
	}

	return nil
}

func sendPacket(p mynet.Packet, conn io.Writer) error {
	return p.Encode(conn)
}

func (s *SensorCollection) logEvent() {

}
