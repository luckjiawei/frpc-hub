package scanner

import (
	"bufio"
	"context"
	"errors"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/afpacket"
	"github.com/google/gopacket/layers"
)

// Device represents a discovered LAN host.
type Device struct {
	IP        string `json:"ip"`
	MAC       string `json:"mac"`
	Hostname  string `json:"hostname"`
	Interface string `json:"interface"`
}

// LocalNetwork describes one scannable subnet on this host.
type LocalNetwork struct {
	Interface string `json:"interface"`
	CIDR      string `json:"cidr"`
	Hosts     int    `json:"hosts"`
}

// PortResult is the result of scanning a single TCP port.
type PortResult struct {
	Port    int    `json:"port"`
	Open    bool   `json:"open"`
	Service string `json:"service"`
	Done    bool   `json:"done,omitempty"` // signals end-of-stream
}

// wellKnownPorts maps port numbers to service names.
var wellKnownPorts = map[int]string{
	21:    "FTP",
	22:    "SSH",
	23:    "Telnet",
	25:    "SMTP",
	53:    "DNS",
	80:    "HTTP",
	110:   "POP3",
	143:   "IMAP",
	443:   "HTTPS",
	445:   "SMB",
	1433:  "MSSQL",
	1521:  "Oracle",
	3306:  "MySQL",
	3389:  "RDP",
	5432:  "PostgreSQL",
	5900:  "VNC",
	6379:  "Redis",
	8080:  "HTTP-Alt",
	8443:  "HTTPS-Alt",
	8888:  "HTTP-Dev",
	9200:  "Elasticsearch",
	11211: "Memcached",
	27017: "MongoDB",
}

// CommonPorts is the default port list for port scanning.
var CommonPorts = func() []int {
	ports := make([]int, 0, len(wellKnownPorts))
	for p := range wellKnownPorts {
		ports = append(ports, p)
	}
	return ports
}()

// hostProbePorts are the ports tried when checking whether a host is alive
// without ARP (TCP fallback mode).
var hostProbePorts = []int{22, 80, 443, 445, 3389, 8080, 8443}

// maxPrefixLen is the smallest prefix we will scan (/20 = 4094 hosts max).
// Subnets larger than this (e.g. Docker /16) are skipped automatically.
const maxPrefixLen = 20

// Service provides LAN discovery and port-scan capabilities.
type Service struct{}

func NewService() *Service { return &Service{} }

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

// GetLocalNetworks returns the subnets that will be scanned.
func (s *Service) GetLocalNetworks() []LocalNetwork {
	var out []LocalNetwork
	for _, iface := range upInterfaces() {
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok || ipNet.IP.To4() == nil || ipNet.IP.IsLoopback() {
				continue
			}
			ones, _ := ipNet.Mask.Size()
			if ones < maxPrefixLen {
				continue
			}
			hosts := (1 << (32 - ones)) - 2
			out = append(out, LocalNetwork{
				Interface: iface.Name,
				CIDR:      ipNet.String(),
				Hosts:     hosts,
			})
		}
	}
	return out
}

// DiscoverDevices scans local subnets for active hosts.
// It first attempts a gopacket/afpacket ARP sweep (requires CAP_NET_RAW).
// If that is unavailable it falls back to TCP-connect probing + reading the
// kernel ARP cache from /proc/net/arp.
func (s *Service) DiscoverDevices(ctx context.Context) ([]Device, error) {
	ifaces := upInterfaces()
	if len(ifaces) == 0 {
		return []Device{}, nil
	}

	// Try ARP sweep on first suitable interface; if permission denied use TCP.
	devices, err := s.arpDiscover(ctx, ifaces)
	if err != nil {
		if isPermissionDenied(err) {
			return s.tcpDiscover(ctx, ifaces)
		}
		return nil, err
	}
	return devices, nil
}

// ScanPorts does a concurrent TCP-connect scan on ip for ports.
// Results are written to ch; the caller must close ch after this returns.
func (s *Service) ScanPorts(ctx context.Context, ip string, ports []int, ch chan<- PortResult) {
	const concurrency = 200
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for _, port := range ports {
		if ctx.Err() != nil {
			break
		}
		sem <- struct{}{}
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			defer func() { <-sem }()

			addr := net.JoinHostPort(ip, strconv.Itoa(p))
			conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
			open := err == nil
			if conn != nil {
				conn.Close()
			}

			result := PortResult{Port: p, Open: open}
			if open {
				result.Service = wellKnownPorts[p]
			}
			select {
			case ch <- result:
			case <-ctx.Done():
			}
		}(port)
	}
	wg.Wait()
}

// ---------------------------------------------------------------------------
// ARP-based discovery (gopacket + afpacket, needs CAP_NET_RAW)
// ---------------------------------------------------------------------------

func (s *Service) arpDiscover(ctx context.Context, ifaces []net.Interface) ([]Device, error) {
	var (
		mu      sync.Mutex
		wg      sync.WaitGroup
		devices = make([]Device, 0)
		seen    = map[string]bool{}
		lastErr error
	)

	for _, iface := range ifaces {
		if len(iface.HardwareAddr) == 0 {
			continue
		}
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok || ipNet.IP.To4() == nil || ipNet.IP.IsLoopback() {
				continue
			}
			ones, _ := ipNet.Mask.Size()
			if ones < maxPrefixLen {
				continue // skip huge subnets like Docker /16
			}

			wg.Add(1)
			go func(iface net.Interface, ip net.IP, network *net.IPNet) {
				defer wg.Done()
				found, err := arpSweep(ctx, iface, ip, network)
				if err != nil {
					mu.Lock()
					lastErr = err
					mu.Unlock()
					return
				}
				mu.Lock()
				for _, d := range found {
					if !seen[d.IP] {
						seen[d.IP] = true
						devices = append(devices, d)
					}
				}
				mu.Unlock()
			}(iface, ipNet.IP, ipNet)
		}
	}

	wg.Wait()

	if len(devices) == 0 && lastErr != nil {
		return nil, lastErr
	}
	return devices, nil
}

func arpSweep(ctx context.Context, iface net.Interface, src net.IP, network *net.IPNet) ([]Device, error) {
	handle, err := afpacket.NewTPacket(
		afpacket.OptInterface(iface.Name),
		afpacket.OptBlockTimeout(5*time.Millisecond),
	)
	if err != nil {
		return nil, err
	}

	sweepCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	results := make(chan Device, 512)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		captureARP(sweepCtx, handle, results)
		close(results)
	}()

	sendARP(sweepCtx, handle, iface, src, network)

	<-sweepCtx.Done()
	handle.Close()
	wg.Wait()

	selfIP := src.String()
	var devices []Device
	seen := map[string]bool{selfIP: true}
	for d := range results {
		if seen[d.IP] {
			continue
		}
		seen[d.IP] = true
		d.Interface = iface.Name
		if names, err := net.LookupAddr(d.IP); err == nil && len(names) > 0 {
			d.Hostname = strings.TrimSuffix(names[0], ".")
		}
		devices = append(devices, d)
	}
	return devices, nil
}

func captureARP(ctx context.Context, handle *afpacket.TPacket, out chan<- Device) {
	pktSrc := gopacket.NewPacketSource(handle, layers.LinkTypeEthernet)
	pktSrc.NoCopy = true
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		pkt, err := pktSrc.NextPacket()
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			continue
		}
		arpLayer := pkt.Layer(layers.LayerTypeARP)
		if arpLayer == nil {
			continue
		}
		arp := arpLayer.(*layers.ARP)
		if arp.Operation != layers.ARPReply {
			continue
		}
		ip := net.IP(arp.SourceProtAddress).String()
		mac := net.HardwareAddr(arp.SourceHwAddress).String()
		select {
		case out <- Device{IP: ip, MAC: mac}:
		default:
		}
	}
}

func sendARP(ctx context.Context, handle *afpacket.TPacket, iface net.Interface, src net.IP, network *net.IPNet) {
	eth := layers.Ethernet{
		SrcMAC:       iface.HardwareAddr,
		DstMAC:       net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		EthernetType: layers.EthernetTypeARP,
	}
	arp := layers.ARP{
		AddrType:          layers.LinkTypeEthernet,
		Protocol:          layers.EthernetTypeIPv4,
		HwAddressSize:     6,
		ProtAddressSize:   4,
		Operation:         layers.ARPRequest,
		SourceHwAddress:   []byte(iface.HardwareAddr),
		SourceProtAddress: []byte(src.To4()),
		DstHwAddress:      []byte{0, 0, 0, 0, 0, 0},
	}
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{FixLengths: true, ComputeChecksums: true}

	for ip := cloneIP(network.IP.Mask(network.Mask)); network.Contains(ip); incrementIP(ip) {
		select {
		case <-ctx.Done():
			return
		default:
		}
		target := make(net.IP, 4)
		copy(target, ip.To4())
		arp.DstProtAddress = target
		if err := gopacket.SerializeLayers(buf, opts, &eth, &arp); err != nil {
			continue
		}
		_ = handle.WritePacketData(buf.Bytes())
		time.Sleep(500 * time.Microsecond)
	}
}

// ---------------------------------------------------------------------------
// TCP-based host discovery fallback (no special privileges required)
// ---------------------------------------------------------------------------

// tcpDiscover probes each host in every local subnet using TCP connect.
// A host is "alive" if any probe port responds (open or connection refused).
// After the sweep it enriches results with MAC addresses from /proc/net/arp.
func (s *Service) tcpDiscover(ctx context.Context, ifaces []net.Interface) ([]Device, error) {
	type result struct {
		ip    string
		iface string
	}

	resultCh := make(chan result, 512)
	var sweepWg sync.WaitGroup

	const hostConcurrency = 100

	for _, iface := range ifaces {
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok || ipNet.IP.To4() == nil || ipNet.IP.IsLoopback() {
				continue
			}
			ones, _ := ipNet.Mask.Size()
			if ones < maxPrefixLen {
				continue
			}

			// Enumerate all host addresses in this subnet.
			var ips []string
			selfIP := ipNet.IP.String()
			for ip := cloneIP(ipNet.IP.Mask(ipNet.Mask)); ipNet.Contains(ip); incrementIP(ip) {
				candidate := ip.String()
				if candidate == selfIP {
					continue
				}
				ips = append(ips, candidate)
			}

			sweepWg.Add(1)
			go func(ifaceName string, ips []string) {
				defer sweepWg.Done()
				sem := make(chan struct{}, hostConcurrency)
				var wg sync.WaitGroup
				for _, ip := range ips {
					if ctx.Err() != nil {
						break
					}
					sem <- struct{}{}
					wg.Add(1)
					go func(ip string) {
						defer wg.Done()
						defer func() { <-sem }()
						if isHostAlive(ctx, ip) {
							select {
							case resultCh <- result{ip: ip, iface: ifaceName}:
							case <-ctx.Done():
							}
						}
					}(ip)
				}
				wg.Wait()
			}(iface.Name, ips)
		}
	}

	// Close resultCh once all sweeps complete.
	go func() {
		sweepWg.Wait()
		close(resultCh)
	}()

	// Collect results.
	arpCache := readARPCache()
	devices := make([]Device, 0)
	seen := map[string]bool{}

	for r := range resultCh {
		if seen[r.ip] {
			continue
		}
		seen[r.ip] = true
		d := Device{IP: r.ip, Interface: r.iface}
		if mac, ok := arpCache[r.ip]; ok {
			d.MAC = mac
		}
		if names, err := net.LookupAddr(r.ip); err == nil && len(names) > 0 {
			d.Hostname = strings.TrimSuffix(names[0], ".")
		}
		devices = append(devices, d)
	}
	return devices, nil
}

// isHostAlive returns true if the host has any TCP port open or actively
// refusing connections (which proves the host is up).
func isHostAlive(ctx context.Context, ip string) bool {
	ctx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
	defer cancel()

	found := make(chan struct{}, 1)
	var wg sync.WaitGroup
	for _, port := range hostProbePorts {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			addr := net.JoinHostPort(ip, strconv.Itoa(p))
			d := net.Dialer{}
			conn, err := d.DialContext(ctx, "tcp", addr)
			if conn != nil {
				conn.Close()
			}
			if err == nil || isConnRefused(err) {
				select {
				case found <- struct{}{}:
				default:
				}
				cancel()
			}
		}(port)
	}
	wg.Wait()
	select {
	case <-found:
		return true
	default:
		return false
	}
}

// isConnRefused checks whether err is ECONNREFUSED (host alive, port closed).
func isConnRefused(err error) bool {
	if err == nil {
		return false
	}
	var opErr *net.OpError
	if !errors.As(err, &opErr) {
		return false
	}
	var syscallErr *os.SyscallError
	if !errors.As(opErr.Err, &syscallErr) {
		return false
	}
	errno, ok := syscallErr.Err.(syscall.Errno)
	return ok && errno == syscall.ECONNREFUSED
}

// readARPCache parses /proc/net/arp and returns an ip→mac map.
func readARPCache() map[string]string {
	f, err := os.Open("/proc/net/arp")
	if err != nil {
		return map[string]string{}
	}
	defer f.Close()

	cache := map[string]string{}
	scanner := bufio.NewScanner(f)
	scanner.Scan() // skip header line
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 4 {
			continue
		}
		ip := fields[0]
		mac := fields[3]
		if mac == "00:00:00:00:00:00" {
			continue
		}
		cache[ip] = mac
	}
	return cache
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func upInterfaces() []net.Interface {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	var out []net.Interface
	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		out = append(out, iface)
	}
	return out
}

func isPermissionDenied(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "permission denied") ||
		strings.Contains(s, "operation not permitted")
}

func cloneIP(ip net.IP) net.IP {
	clone := make(net.IP, len(ip))
	copy(clone, ip)
	return clone
}

func incrementIP(ip net.IP) {
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] != 0 {
			break
		}
	}
}
