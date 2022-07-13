package main

import (
	mynet "github.com/Heanthor/quill-secure/net"
	"github.com/Heanthor/quill-secure/node/sensor"
	"github.com/rs/zerolog/log"
	"net"
	"time"
)

type SensorCollection struct {
	deviceID      uint8
	activeSensors []sensor.Sensor

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
		pingTicker: t,
	}
}

func (s *SensorCollection) RegisterSensors() {
	f := sensor.FakeSensor{Buf: "fake data"}
	if f.Ping() == nil {
		log.Info().Msg("Registered fake sensor")
		s.activeSensors = append(s.activeSensors, &f)
	}
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
	log.Debug().Msg("pingLeader")
	conn, err := net.DialTCP("tcp", nil, &net.TCPAddr{
		IP:   net.ParseIP(s.leader.Host),
		Port: s.leader.Port,
	})
	if err != nil {
		log.Err(err).Msg("Error creating TCP conn to leader")
		return err
	}
	defer conn.Close()

	if _, err := conn.Write([]byte("ping")); err != nil {
		log.Err(err).Msg("Error pinging leader")
		return err
	}

	return nil
}

func (s *SensorCollection) Poll() {
	for {
		for _, sn := range s.activeSensors {
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
}

func (s *SensorCollection) SendPacket(p mynet.Packet) error {
	return nil
}
