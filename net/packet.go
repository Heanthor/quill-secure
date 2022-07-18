package net

import (
	"encoding/gob"
	"errors"
	"io"
	"net"
)

const (
	PacketTypeAnnounce = iota + 1
	PacketTypeSensorData
)

type Packet struct {
	// UID is the "deviceID" of the node sending the packet
	UID uint8
	// Typ is the type of packet
	Typ  uint8
	Data interface{}
}

type Dest struct {
	Host net.IP
	Port int
}

func ParseHost(host string) (net.IP, error) {
	ip := net.ParseIP(host)
	if ip == nil && host != "localhost" {
		return nil, errors.New("invalid host")
	}

	return ip, nil
}

func (p Packet) Encode(w io.Writer) error {
	e := gob.NewEncoder(w)

	return e.Encode(p)
}

// ReadPacket attempts to stream data off the reader and convert it into a Packet
func ReadPacket(r io.Reader) (*Packet, error) {
	var p Packet
	err := gob.NewDecoder(r).Decode(&p)

	if err != nil {
		return nil, err
	}

	return &p, nil
}
