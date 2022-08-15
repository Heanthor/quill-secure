package main

import (
	"fmt"
	"github.com/Heanthor/quill-secure/db"
	mynet "github.com/Heanthor/quill-secure/net"
	"github.com/Heanthor/quill-secure/node/sensor"
	"github.com/rs/zerolog/log"
	"net"
	"strconv"
	"time"
)

const (
	ConnType = "tcp"
)

type LeaderNet struct {
	DB *db.DB

	dest                mynet.Dest
	listener            net.Listener
	activeNodes         map[uint8]remoteNode
	nodePingTimeoutSecs int

	closing bool

	datapoints chan SensorData
}

type remoteNode struct {
	DeviceID uint8
	// TODO this should be a list
	SensorType uint8
	Active     bool
	LastSeenAt time.Time
}

type SensorData struct {
	sensor remoteNode
	data   sensor.Data
}

// NewLeaderNet returns a new LeaderNet with listener initialized on host and port
func NewLeaderNet(host string, port, nodePingTimeoutSecs int, db *db.DB) (*LeaderNet, error) {
	listener, err := net.Listen(ConnType, host+":"+strconv.Itoa(port))
	if err != nil {
		return nil, fmt.Errorf("NewLeaderNet error starting listener: %w", err)
	}
	ip, err := mynet.ParseHost(host)
	if err != nil {
		log.Fatal().Msg("Invalid host parameter")
	}

	return &LeaderNet{
		DB: db,
		dest: mynet.Dest{
			Host: ip,
			Port: port,
		},
		listener:            listener,
		datapoints:          make(chan SensorData, 100),
		activeNodes:         make(map[uint8]remoteNode),
		nodePingTimeoutSecs: nodePingTimeoutSecs,
	}, nil
}

func (l *LeaderNet) StartListening() error {
	log.Info().Str("host", l.dest.Host.String()).Int("port", l.dest.Port).Msg("Started listening")
	go l.nodeAgeWorker()
	go l.sensorReadoutConsumerWorker()

	for {
		if l.closing {
			return nil
		}
		// Listen for an incoming connection.
		conn, err := l.listener.Accept()
		if err != nil && !l.closing {
			log.Err(err).Msg("StartListening: Error accepting connection")
			continue
		}

		// TODO a thread pool should be good here
		go l.handleRequest(conn)
	}
}

// nodeAgeWorker blocks and checks time since last ping from all connected sensors, marking inactive if they have gone silent
func (l *LeaderNet) nodeAgeWorker() {
	t := time.NewTicker(time.Second)
	for ts := range t.C {
		for _, sn := range l.activeNodes {
			if ts.Sub(sn.LastSeenAt) > (time.Second*time.Duration(l.nodePingTimeoutSecs)) && sn.Active {
				log.Warn().Uint8("deviceID", sn.DeviceID).Msg("Node has gone offline, marking inactive")
				sn.Active = false
				l.activeNodes[sn.DeviceID] = sn
			}
		}
	}
}

func (l *LeaderNet) sensorReadoutConsumerWorker() {
	for {
		sd := <-l.datapoints
		switch sd.data.Typ {
		case sensor.TypeFake:
			log.Debug().Msg("Parse fake sensor data")
		case sensor.TypeAtmospheric:
			ad := sensor.ParseValidAtmosphericSensorLine(string(sd.data.Data))
			if err := l.DB.RecordAtmosphericMeasurement(ad); err != nil {
				log.Err(err).Msg("error recording atmospheric measurement")
			}
		}
	}
}

func (l *LeaderNet) handleRequest(conn net.Conn) {
	p, err := mynet.ReadPacket(conn)
	if err != nil {
		log.Err(err).Msg("handleRequest: Error decoding packet")
		return
	}
	defer conn.Close()

	l.parseIncomingPacket(p)
}

func (l *LeaderNet) parseIncomingPacket(p *mynet.Packet) {
	switch p.Typ {
	case mynet.PacketTypeAnnounce:
		l.nodeAnnounce(p)
	case mynet.PacketTypeSensorData:
		log.Debug().Uint8("deviceID", p.UID).Msg("sensor readout")
		l.datapoints <- SensorData{
			sensor: remoteNode{
				DeviceID:   p.UID,
				SensorType: p.Typ,
			},
			data: p.Data.(sensor.Data),
		}
	}
}

func (l *LeaderNet) nodeAnnounce(p *mynet.Packet) {
	if entry, ok := l.activeNodes[p.UID]; !ok {
		log.Info().Uint8("deviceID", p.UID).Str("type", sensor.NameByType(int(p.Typ))).Msg("New node connected")
		l.activeNodes[p.UID] = remoteNode{
			DeviceID:   p.UID,
			SensorType: p.Typ,
			Active:     true,
			LastSeenAt: time.Now(),
		}
	} else {
		entry.LastSeenAt = time.Now()
		if !entry.Active {
			log.Info().
				Uint8("deviceID", p.UID).
				Str("type", sensor.NameByType(int(p.Typ))).
				Msg("Previously seen node reconnected")
			entry.Active = true
		}

		l.activeNodes[p.UID] = entry
	}
}

func (l *LeaderNet) Close() {
	l.closing = true
	log.Debug().Msg("Close listener")
	if l.listener != nil {
		l.listener.Close()
	}
}
