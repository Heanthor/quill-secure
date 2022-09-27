package net

import (
	"fmt"
	"github.com/Heanthor/quill-secure/db"
	mynet "github.com/Heanthor/quill-secure/net"
	"github.com/Heanthor/quill-secure/node/sensor"
	"github.com/rs/zerolog/log"
	"net"
	"strconv"
	"sync"
	"time"
)

const (
	ConnType = "tcp"
)

type LeaderNet struct {
	DB *db.DB

	dest                mynet.Dest
	listener            net.Listener
	seenNodes           map[uint8]remoteNode
	nodePingTimeoutSecs int

	closing bool

	datapoints chan SensorData
	nodeLock   sync.Mutex
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
func NewLeaderNet(port, nodePingTimeoutSecs int, db *db.DB) (*LeaderNet, error) {
	listener, err := net.Listen(ConnType, ":"+strconv.Itoa(port))
	if err != nil {
		return nil, fmt.Errorf("NewLeaderNet error starting listener: %w", err)
	}

	return &LeaderNet{
		DB: db,
		dest: mynet.Dest{
			Port: port,
		},
		listener:            listener,
		datapoints:          make(chan SensorData, 100),
		seenNodes:           make(map[uint8]remoteNode),
		nodePingTimeoutSecs: nodePingTimeoutSecs,
	}, nil
}

func (l *LeaderNet) StartListening() error {
	log.Info().Int("port", l.dest.Port).Msg("Started listening")
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

type ActiveNodesFunc func() int

// ActiveNodesFunc returns a function which counts the number of nodes currently connected and sending data to leader.
// If a node has previously connected, but is not currently active in sending data, it is not counted.
func (l *LeaderNet) ActiveNodesFunc() ActiveNodesFunc {
	return func() int {
		l.nodeLock.Lock()
		defer l.nodeLock.Unlock()

		count := 0
		for _, n := range l.seenNodes {
			if n.Active {
				count++
			}
		}

		return count
	}
}

// nodeAgeWorker blocks and checks time since last ping from all connected sensors, marking inactive if they have gone silent
func (l *LeaderNet) nodeAgeWorker() {
	t := time.NewTicker(time.Second)
	for ts := range t.C {
		l.nodeLock.Lock()
		for _, sn := range l.seenNodes {
			if ts.Sub(sn.LastSeenAt) > (time.Second*time.Duration(l.nodePingTimeoutSecs)) && sn.Active {
				log.Warn().Uint8("deviceID", sn.DeviceID).Msg("Node has gone offline, marking inactive")
				sn.Active = false
				l.seenNodes[sn.DeviceID] = sn
			}
		}
		l.nodeLock.Unlock()
	}
}

// sensorReadoutConsumerWorker listens on l.datapoints and saves/actions on incoming sensor data.
func (l *LeaderNet) sensorReadoutConsumerWorker() {
	for {
		sd := <-l.datapoints
		switch sd.data.Typ {
		case sensor.TypeFake:
			log.Debug().Msg("Parse fake sensor data")
		case sensor.TypeAtmospheric:
			ad := sensor.ParseValidAtmosphericSensorLine(string(sd.data.Data))
			if err := l.DB.RecordAtmosphericMeasurement(ad, sd.sensor.DeviceID); err != nil {
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

// nodeAnnounce handles a node announce packet. This is a periodic ping from each node
// which signals continued connection with the leader.
func (l *LeaderNet) nodeAnnounce(p *mynet.Packet) {
	l.nodeLock.Lock()
	defer l.nodeLock.Unlock()
	if entry, ok := l.seenNodes[p.UID]; !ok {
		log.Info().Uint8("deviceID", p.UID).Msg("New node connected")
		l.seenNodes[p.UID] = remoteNode{
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
				Msg("Previously seen node reconnected")
			entry.Active = true
		}

		l.seenNodes[p.UID] = entry
	}
}

func (l *LeaderNet) Close() {
	l.closing = true
	log.Debug().Msg("Close listener")
	if l.listener != nil {
		l.listener.Close()
	}
}
