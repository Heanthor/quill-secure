package main

import (
	"github.com/rs/zerolog/log"
	"net"
	"strconv"
)

const (
	ConnType = "tcp"
)

type LeaderNet struct {
	host string
	port int
}

func (l *LeaderNet) StartListening() error {
	listener, err := net.Listen(ConnType, l.host+":"+strconv.Itoa(l.port))
	if err != nil {
		log.Err(err).Msg("Error starting listener")
		return err
	}
	// Close the listener when the application closes.
	defer listener.Close()
	log.Info().Str("host", l.host).Int("port", l.port).Msg("Started listening")
	for {
		// Listen for an incoming connection.
		conn, err := listener.Accept()
		if err != nil {
			log.Err(err).Msg("Error accepting connection")
			continue
		}
		// Handle connections in a new goroutine.
		go l.handleRequest(conn)
	}
}

func (l *LeaderNet) handleRequest(conn net.Conn) {
	log.Info().Msg("Handling request")
}

func (l *LeaderNet) Close() {

}
