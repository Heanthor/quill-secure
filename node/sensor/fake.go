package sensor

import (
	"github.com/rs/zerolog/log"
	"time"
)

type FakeSensor struct {
	Buf string
}

func (f *FakeSensor) Type() uint8 {
	return TypeFake
}

func (f *FakeSensor) Init() error {
	return nil
}

func (f *FakeSensor) TypeStr() string {
	return NameByType(int(f.Type()))
}

func (f *FakeSensor) Ping() error {
	return nil
}

func (f *FakeSensor) Data() (chan Data, chan error) {
	errCh := make(chan error)
	dataCh := make(chan Data)

	t := time.NewTicker(5 * time.Second)
	go func() {
		for {
			<-t.C
			dataCh <- Data{
				Typ:  f.Type(),
				Data: []byte(f.Buf),
			}
		}
	}()

	return dataCh, errCh
}

func (f *FakeSensor) Close() {
	log.Info().Msg("close fake sensor")
}
