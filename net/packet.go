package net

import (
	"encoding/gob"
	"io"
)

type Packet struct {
	UID  uint8
	Typ  uint8
	Data []byte
}

type Dest struct {
	Host string
	Port int
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
