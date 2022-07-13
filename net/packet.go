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
