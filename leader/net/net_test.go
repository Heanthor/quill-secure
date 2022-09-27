package net

import (
	mynet "github.com/Heanthor/quill-secure/net"
	"github.com/Heanthor/quill-secure/node/sensor"
	"testing"
	"time"
)

func TestLeaderNet_nodeAnnounce(t *testing.T) {
	time1, _ := time.Parse(time.RFC1123, "Sun, 17 Jul 2022 22:13:37 GMT")
	deviceID := uint8(1)
	type fields struct {
		activeSensors map[uint8]remoteNode
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
			fields: fields{activeSensors: map[uint8]remoteNode{
				deviceID: {
					DeviceID:   deviceID,
					SensorType: sensor.TypeFake,
					LastSeenAt: time1,
					Active:     true,
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
				seenNodes: tt.fields.activeSensors,
			}
			l.nodeAnnounce(tt.args.p)
			if l.seenNodes[deviceID].LastSeenAt == time1 {
				t.Fatalf("time not updated")
			}
		})
	}
}
