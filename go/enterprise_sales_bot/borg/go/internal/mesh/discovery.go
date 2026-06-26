package mesh

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"time"
)

// DiscoveryConfig controls A2A mesh discovery behavior.
type DiscoveryConfig struct {
	// Port is the UDP port for discovery broadcasts.
	Port int
	// Interval is how often beacon packets are sent.
	Interval time.Duration
	// Timeout is how long before a peer is considered lost.
	Timeout time.Duration
	// InterfaceName restricts discovery to a specific network interface (empty = all).
	InterfaceName string
	// Labels are arbitrary key-value pairs broadcast with this node.
	Labels map[string]string
}

// DefaultDiscoveryConfig returns sensible defaults for local network discovery.
func DefaultDiscoveryConfig() DiscoveryConfig {
	return DiscoveryConfig{
		Port:      4301,
		Interval:  5 * time.Second,
		Timeout:   30 * time.Second,
		Labels:    map[string]string{},
	}
}

// BeaconPacket is the JSON payload broadcast over UDP.
type BeaconPacket struct {
	NodeID    string            `json:"node_id"`
	Addr      string            `json:"addr"`
	Port      int               `json:"port"`
	Timestamp int64             `json:"ts"`
	Caps      []string          `json:"caps,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// DiscoveredPeer represents a peer found on the local network.
type DiscoveredPeer struct {
	NodeID    string
	Addr      string
	Port      int
	LastSeen  time.Time
	Caps      []string
	Labels    map[string]string
}

// DiscoveryService provides UDP-based local network peer discovery.
type DiscoveryService struct {
	cfg      DiscoveryConfig
	nodeID   string
	httpPort int
	localCap []string
	peers    map[string]*DiscoveredPeer
	stopCh   chan struct{}
}

// NewDiscoveryService creates a new discovery service.
func NewDiscoveryService(nodeID string, httpPort int, localCap []string, cfg DiscoveryConfig) *DiscoveryService {
	if cfg.Port == 0 {
		cfg = DefaultDiscoveryConfig()
	}
	return &DiscoveryService{
		cfg:      cfg,
		nodeID:   nodeID,
		httpPort: httpPort,
		localCap: localCap,
		peers:    make(map[string]*DiscoveredPeer),
		stopCh:   make(chan struct{}),
	}
}

// Start begins broadcasting beacon packets and listening for peers.
func (d *DiscoveryService) Start(ctx context.Context) error {
	// Get broadcast address
	broadcastAddr := fmt.Sprintf("255.255.255.255:%d", d.cfg.Port)

	// Resolve UDP address
	addr, err := net.ResolveUDPAddr("udp", broadcastAddr)
	if err != nil {
		return fmt.Errorf("mesh discovery: resolve broadcast addr: %w", err)
	}

	// Create UDP connection for sending
	sendConn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return fmt.Errorf("mesh discovery: dial broadcast: %w", err)
	}

	// Create UDP connection for listening
	listenAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", d.cfg.Port))
	if err != nil {
		sendConn.Close()
		return fmt.Errorf("mesh discovery: resolve listen addr: %w", err)
	}

	recvConn, err := net.ListenUDP("udp", listenAddr)
	if err != nil {
		sendConn.Close()
		return fmt.Errorf("mesh discovery: listen: %w", err)
	}

	// Get local IP for the beacon packet
	localIP := d.getLocalIP()

	go d.broadcastLoop(ctx, sendConn, localIP)
	go d.listenLoop(ctx, recvConn)
	go d.evictionLoop(ctx)

	return nil
}

// Stop halts the discovery service.
func (d *DiscoveryService) Stop() {
	close(d.stopCh)
}

// Peers returns all currently known peers.
func (d *DiscoveryService) Peers() []DiscoveredPeer {
	result := make([]DiscoveredPeer, 0, len(d.peers))
	for _, peer := range d.peers {
		result = append(result, *peer)
	}
	return result
}

// PeerByCapability returns peers that have all the specified capabilities.
func (d *DiscoveryService) PeerByCapability(required []string) []DiscoveredPeer {
	var result []DiscoveredPeer
	for _, peer := range d.peers {
		if containsAll(peer.Caps, required) {
			result = append(result, *peer)
		}
	}
	return result
}

// PeerCount returns the number of known peers.
func (d *DiscoveryService) PeerCount() int {
	return len(d.peers)
}

func (d *DiscoveryService) broadcastLoop(ctx context.Context, conn *net.UDPConn, localIP string) {
	defer conn.Close()

	ticker := time.NewTicker(d.cfg.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-d.stopCh:
			return
		case <-ticker.C:
			pkt := BeaconPacket{
				NodeID:    d.nodeID,
				Addr:      localIP,
				Port:      d.httpPort,
				Timestamp: time.Now().UnixMilli(),
				Caps:      d.localCap,
				Labels:    d.cfg.Labels,
			}
			data, err := json.Marshal(pkt)
			if err != nil {
				continue
			}
			conn.Write(data)
		}
	}
}

func (d *DiscoveryService) listenLoop(ctx context.Context, conn *net.UDPConn) {
	defer conn.Close()

	buf := make([]byte, 4096)

	for {
		select {
		case <-ctx.Done():
			return
		case <-d.stopCh:
			return
		default:
		}

		// Set a read deadline so we can check ctx periodically
		conn.SetReadDeadline(time.Now().Add(1 * time.Second))

		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			if ctx.Err() != nil {
				return
			}
			continue
		}

		var pkt BeaconPacket
		if err := json.Unmarshal(buf[:n], &pkt); err != nil {
			continue
		}

		// Ignore our own beacons
		if pkt.NodeID == d.nodeID {
			continue
		}

		// Update or add peer
		peer, exists := d.peers[pkt.NodeID]
		if !exists {
			peer = &DiscoveredPeer{
				NodeID: pkt.NodeID,
			}
			d.peers[pkt.NodeID] = peer
		}
		peer.Addr = pkt.Addr
		peer.Port = pkt.Port
		peer.LastSeen = time.Now()
		peer.Caps = pkt.Caps
		if pkt.Labels != nil {
			peer.Labels = pkt.Labels
		}
	}
}

func (d *DiscoveryService) evictionLoop(ctx context.Context) {
	ticker := time.NewTicker(d.cfg.Interval * 2)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-d.stopCh:
			return
		case <-ticker.C:
			now := time.Now()
			for nodeID, peer := range d.peers {
				if now.Sub(peer.LastSeen) > d.cfg.Timeout {
					delete(d.peers, nodeID)
				}
			}
		}
	}
}

func (d *DiscoveryService) getLocalIP() string {
	// Try to find a non-loopback local IP
	interfaces, err := net.Interfaces()
	if err != nil {
		return "127.0.0.1"
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		// If specific interface requested, skip others
		if d.cfg.InterfaceName != "" && iface.Name != d.cfg.InterfaceName {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			if ipNet.IP.IsLoopback() {
				continue
			}
			if ipv4 := ipNet.IP.To4(); ipv4 != nil {
				return ipv4.String()
			}
		}
	}

	return "127.0.0.1"
}

// BeaconAddr returns the UDP address where beacons are sent.
func (d *DiscoveryService) BeaconAddr() string {
	return "255.255.255.255:" + strconv.Itoa(d.cfg.Port)
}

// ParseBeaconAddr extracts host and port from a peer's Addr:Port.
func ParseBeaconAddr(addr string, port int) string {
	return net.JoinHostPort(addr, strconv.Itoa(port))
}
