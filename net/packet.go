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

func (p Packet) Encode(w io.Writer) error {
	e := gob.NewEncoder(w)

	return e.Encode(p)
}
