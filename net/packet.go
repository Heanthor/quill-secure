package net

import (
	"encoding/gob"
	"io"
)

type Packet struct {
	UID uint64
	Typ uint8
}

func (p Packet) Encode(w io.Writer) error {
	e := gob.NewEncoder(w)

	return e.Encode(p)
}
