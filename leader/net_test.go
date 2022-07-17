package main

import (
	mynet "github.com/Heanthor/quill-secure/net"
	"github.com/Heanthor/quill-secure/node/sensor"
	"testing"
	"time"
)

func TestLeaderNet_handleSensorPing(t *testing.T) {
	time1, _ := time.Parse(time.RFC1123, "Sun, 17 Jul 2022 22:13:37 GMT")
	deviceID := uint8(1)
	type fields struct {
		activeSensors map[uint8]remoteSensor
	}
	type args struct {
		p *mynet.Packet
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "updates LastSeenAt for returning sensor",
			fields: fields{activeSensors: map[uint8]remoteSensor{
				deviceID: {
					DeviceID:   deviceID,
					Type:       sensor.TypeFake,
					LastSeenAt: time1,
				},
			}},
			args: args{p: &mynet.Packet{
				UID: deviceID,
				Typ: sensor.TypeFake,
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &LeaderNet{
				activeSensors: tt.fields.activeSensors,
			}
			l.handleSensorPing(tt.args.p)
			if l.activeSensors[deviceID].LastSeenAt == time1 {
				t.Fatalf("time not updated")
			}
		})
	}
}
