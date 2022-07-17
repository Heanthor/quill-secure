package main

import (
	mynet "github.com/Heanthor/quill-secure/net"
	"github.com/Heanthor/quill-secure/node/sensor"
	"github.com/rs/zerolog/log"
	"net"
	"strconv"
)

const (
	ConnType = "tcp"
)

type LeaderNet struct {
	dest     mynet.Dest
	listener net.Listener

	datapoints <-chan SensorData
}

type SensorData struct {
	deviceID uint8
	data     []byte
}

// NewLeaderNet returns a new LeaderNet with listener initialized on host and port
func NewLeaderNet(host string, port int) (*LeaderNet, error) {
	listener, err := net.Listen(ConnType, host+":"+strconv.Itoa(port))
	if err != nil {
		log.Err(err).Msg("Error starting listener")
		return nil, err
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
		listener:   listener,
		datapoints: make(chan SensorData),
	}, nil
}

func (l *LeaderNet) StartListening() error {
	log.Info().Str("host", l.dest.Host.String()).Int("port", l.dest.Port).Msg("Started listening")
	for {
		// Listen for an incoming connection.
		conn, err := l.listener.Accept()
		if err != nil {
			log.Err(err).Msg("Error accepting connection")
			continue
		}
		// Handle connections in a new goroutine.
		go l.handleRequest(conn)
	}
}

func (l *LeaderNet) handleRequest(conn net.Conn) {
	p, err := mynet.ReadPacket(conn)
	if err != nil {
		log.Err(err).Msg("Error decoding packet")
		return
	}
	// TODO 2:23PM ERR Error accepting connection error="accept tcp 127.0.0.1:5530: use of closed network connection"
	defer conn.Close()

	switch p.Typ {
	case 0:
		log.Debug().Uint8("deviceID", p.UID).Msg("ping")
	case sensor.TypeFake:
		log.Info().Uint8("deviceID", p.UID).Msg("fake sensor readout")
	}
}

func (l *LeaderNet) Close() {
	log.Debug().Msg("Close listener")
	if l.listener != nil {
		l.listener.Close()
	}
}
