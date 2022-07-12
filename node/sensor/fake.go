package sensor

type FakeSensor struct {
	Buf string
}

func (f *FakeSensor) Type() uint8 {
	return TypeFake
}

func (f *FakeSensor) Ping() error {
	return nil
}

func (f *FakeSensor) Poll() ([]byte, error) {
	return []byte(f.Buf), nil
}

func (f *FakeSensor) Close() {
}
