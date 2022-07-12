package main

import (
	"github.com/Heanthor/quill-secure/net"
	"github.com/Heanthor/quill-secure/node/sensor"
	"github.com/rs/zerolog/log"
)

type SensorCollection struct {
	deviceID      uint8
	activeSensors []sensor.Sensor
}

func (s *SensorCollection) RegisterSensors() {
	f := sensor.FakeSensor{Buf: "fake data"}
	if f.Ping() == nil {
		log.Info().Msg("Registered fake sensor")
		s.activeSensors = append(s.activeSensors, &f)
	}
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
			p := net.Packet{
				UID:  s.deviceID,
				Typ:  sn.Type(),
				Data: data,
			}

		}
	}
}
