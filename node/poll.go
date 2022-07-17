package main

import (
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
	sensorPings   chan sensor.Sensor

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
		sensorPings: make(chan sensor.Sensor),
		doneChan:    make(chan bool),
		sendChan:    sendChan,
		pingTicker:  t,
	}
}

func (s *SensorCollection) RegisterSensors() {
	f := sensor.FakeSensor{Buf: "fake data"}
	if f.Ping() == nil {
		s.registerSensor(&f)
	}
	// TODO more sensors here
}

func (s *SensorCollection) registerSensor(sn sensor.Sensor) {
	log.Info().Str("type", sn.TypeStr()).Msg("Registered new sensor")

	t := time.NewTicker(time.Duration(sn.PollRate()) * time.Millisecond)
	go func() {
		for range t.C {
			s.sensorPings <- sn
		}
	}()
	s.activeSensors = append(s.activeSensors, sn)
}

func (s *SensorCollection) StopPolling() {
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
					log.Error().Msg("Cannot reach leader node")
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
		data, err := sn.Poll()
		if err != nil {
			log.Err(err).Uint8("sensor", sn.Type()).Msg("Error in sensor")
			continue
		}

		// send data to leader
		p := mynet.Packet{
			UID:  s.deviceID,
			Typ:  sn.Type(),
			Data: data,
		}

		select {
		case s.sendChan <- outgoingPacketWrapper{p, sn}:
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
		log.Err(err).Msg("Error creating TCP conn to leader")
		return err
	}
	defer conn.Close()

	if err := sendPacket(p, conn); err != nil {
		log.Err(err).Msg("Error sending packet")
		return err
	}

	return nil
}

func sendPacket(p mynet.Packet, conn io.Writer) error {
	return p.Encode(conn)
}

func (s *SensorCollection) logEvent() {

}
