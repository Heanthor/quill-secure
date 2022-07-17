package main

import (
	"fmt"
	mynet "github.com/Heanthor/quill-secure/net"
	"github.com/Heanthor/quill-secure/node/sensor"
	"github.com/rs/zerolog/log"
	"net"
	"strconv"
	"time"
)

const (
	ConnType = "tcp"
)

type LeaderNet struct {
	dest          mynet.Dest
	listener      net.Listener
	activeSensors map[uint8]remoteSensor

	closing bool

	datapoints chan SensorData
}

type remoteSensor struct {
	DeviceID   uint8
	Type       uint8
	LastSeenAt time.Time
}

type SensorData struct {
	sensor remoteSensor
	data   []byte
}

// NewLeaderNet returns a new LeaderNet with listener initialized on host and port
func NewLeaderNet(host string, port int) (*LeaderNet, error) {
	listener, err := net.Listen(ConnType, host+":"+strconv.Itoa(port))
	if err != nil {
		return nil, fmt.Errorf("NewLeaderNet error starting listener: %w", err)
	}
	ip, err := mynet.ParseHost(host)
	if err != nil {
		log.Fatal().Msg("Invalid host parameter")
	}

	return &LeaderNet{
		dest: mynet.Dest{
			Host: ip,
			Port: port,
		},
		listener:      listener,
		datapoints:    make(chan SensorData),
		activeSensors: make(map[uint8]remoteSensor),
	}, nil
}

func (l *LeaderNet) StartListening() error {
	log.Info().Str("host", l.dest.Host.String()).Int("port", l.dest.Port).Msg("Started listening")
	for {
		if l.closing {
			return nil
		}
		// Listen for an incoming connection.
		conn, err := l.listener.Accept()
		if err != nil && !l.closing {
			log.Err(err).Msg("StartListening: Error accepting connection")
			continue
		}

		// TODO a thread pool should be good here
		go l.handleRequest(conn)
	}
}

func (l *LeaderNet) handleRequest(conn net.Conn) {
	p, err := mynet.ReadPacket(conn)
	if err != nil {
		log.Err(err).Msg("handleRequest: Error decoding packet")
		return
	}
	defer conn.Close()

	switch p.Typ {
	case 0:
		// TODO age out sensors if they have not been seen in a ping for x minutes
		l.handleSensorPing(p)
	case sensor.TypeFake:
		log.Info().Uint8("deviceID", p.UID).Msg("fake sensor readout")
		l.datapoints <- SensorData{
			sensor: remoteSensor{
				DeviceID: p.UID,
				Type:     p.Typ,
			},
			data: p.Data,
		}
	}
}

func (l *LeaderNet) handleSensorPing(p *mynet.Packet) {
	log.Debug().Uint8("deviceID", p.UID).Msg("ping")

	if entry, ok := l.activeSensors[p.UID]; !ok {
		log.Info().Uint8("deviceID", p.UID).Str("type", sensor.NameByType(int(p.Typ))).Msg("New sensor connected")
		l.activeSensors[p.UID] = remoteSensor{
			DeviceID:   p.UID,
			Type:       p.Typ,
			LastSeenAt: time.Now(),
		}
	} else {
		entry.LastSeenAt = time.Now()
		l.activeSensors[p.UID] = entry
	}
}

func (l *LeaderNet) Close() {
	l.closing = true
	log.Debug().Msg("Close listener")
	if l.listener != nil {
		l.listener.Close()
	}
}
