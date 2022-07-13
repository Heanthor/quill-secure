package main

import (
	mynet "github.com/Heanthor/quill-secure/net"
	"github.com/rs/zerolog/log"
	"net"
	"strconv"
)

const (
	ConnType = "tcp"
)

type LeaderNet struct {
	dest mynet.Dest

	listener net.Listener
}

func (l *LeaderNet) StartListening() error {
	listener, err := net.Listen(ConnType, l.dest.Host+":"+strconv.Itoa(l.dest.Port))
	if err != nil {
		log.Err(err).Msg("Error starting listener")
		return err
	}

	log.Info().Str("host", l.dest.Host).Int("port", l.dest.Port).Msg("Started listening")
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
	log.Debug().Msg("Close listener")
	if l.listener != nil {
		l.listener.Close()
	}
}
