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

	leader     mynet.Dest
	pingTicker *time.Ticker
}

func NewSensorCollection(deviceID uint8, host string, port, pingIntervalSecs int) SensorCollection {
	t := time.NewTicker(time.Duration(pingIntervalSecs) * time.Second)

	return SensorCollection{
		deviceID: deviceID,
		leader: mynet.Dest{
			Host: host,
			Port: port,
		},
		sensorPings: make(chan sensor.Sensor),
		pingTicker:  t,
	}
}

func (s *SensorCollection) RegisterSensors() {
	f := sensor.FakeSensor{Buf: "fake data"}
	if f.Ping() == nil {
		log.Info().Msg("Registered fake sensor")
		t := time.NewTicker(time.Duration(f.PollRate()) * time.Millisecond)
		go func() {
			for range t.C {
				s.sensorPings <- &f
			}
		}()
		s.activeSensors = append(s.activeSensors, &f)
	}
}

func (s *SensorCollection) StopPolling() {
	close(s.sensorPings)
}

// startPing pings the leader node every pingIntervalSecs seconds
func (s *SensorCollection) startPing() {
	go func() {
		for {
			select {
			case <-s.pingTicker.C:
				if err := s.pingLeader(); err != nil {
					// todo do something else here
					log.Err(err).Msg("Error pinging leader")
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

func (s *SensorCollection) Poll() {
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

		// todo this shouldn't block the other readouts
		if err := s.SendPacket(p); err != nil {
			log.Err(err).Uint8("sensor", sn.Type()).Msg("Error sending packet")
			continue
		}
	}
}

// SendPacket opens a TCP conn to leader, sends the packet, and closes the connection
func (s *SensorCollection) SendPacket(p mynet.Packet) error {
	log.Debug().Interface("packet", p).Msg("SendPacket")
	conn, err := net.DialTCP("tcp", nil, &net.TCPAddr{
		IP:   net.ParseIP(s.leader.Host),
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
