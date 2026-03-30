package icmp

import (
	"context"
	"net"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// ProbeResult holds the outcome of a single ICMP echo probe.
type ProbeResult struct {
	Reachable bool
	RTTMs     int    // Round-trip time in milliseconds; 0 when unreachable
	Error     string // Human-readable error; empty when reachable
}

// Probe sends a single ICMP echo request to host and reports reachability.
// timeout controls the maximum time to wait for a reply.
//
// The function:
//   - Resolves host to an IPv4 address before probing.
//   - Tries unprivileged datagram ICMP first (Linux ping_group_range / macOS).
//   - Falls back to raw IP socket (requires root or CAP_NET_RAW).
//   - Returns an error result rather than panicking when capability is absent.
//   - IPv6 hosts are currently unsupported and return a graceful error.
func Probe(host string, timeout time.Duration) ProbeResult {
	if host == "" {
		return ProbeResult{Error: "host must not be empty"}
	}

	if timeout <= 0 {
		timeout = 2 * time.Second
	}

	// Resolve to IPv4 address
	ips, err := net.DefaultResolver.LookupHost(context.Background(), host)
	if err != nil || len(ips) == 0 {
		msg := "DNS resolution failed"
		if err != nil {
			msg = err.Error()
		}
		return ProbeResult{Error: msg}
	}

	// Pick first IPv4 address
	var target string
	for _, ip := range ips {
		if net.ParseIP(ip).To4() != nil {
			target = ip
			break
		}
	}
	if target == "" {
		return ProbeResult{Error: "no IPv4 address resolved for host (IPv6-only hosts are not supported in H2)"}
	}

	rtt, probeErr := sendICMPEcho(target, timeout)
	if probeErr != nil {
		return ProbeResult{Error: probeErr.Error()}
	}

	return ProbeResult{
		Reachable: true,
		RTTMs:     int(rtt.Milliseconds()),
	}
}

// sendICMPEcho sends a single ICMP echo request and waits for a reply within timeout.
func sendICMPEcho(target string, timeout time.Duration) (time.Duration, error) {
	// Try unprivileged first (Linux with ping_group_range or macOS).
	conn, err := icmp.ListenPacket("udp4", "0.0.0.0")
	network := "udp4"
	if err != nil {
		// Fall back to privileged raw socket.
		conn, err = icmp.ListenPacket("ip4:icmp", "0.0.0.0")
		network = "ip4:icmp"
		if err != nil {
			return 0, err
		}
	}
	defer conn.Close()

	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   1,
			Seq:  1,
			Data: []byte("ogoune-probe"),
		},
	}
	wb, err := msg.Marshal(nil)
	if err != nil {
		return 0, err
	}

	deadline := time.Now().Add(timeout)
	if err := conn.SetDeadline(deadline); err != nil {
		return 0, err
	}

	var dst net.Addr
	if network == "udp4" {
		dst = &net.UDPAddr{IP: net.ParseIP(target)}
	} else {
		dst = &net.IPAddr{IP: net.ParseIP(target)}
	}

	start := time.Now()
	if _, err := conn.WriteTo(wb, dst); err != nil {
		return 0, err
	}

	rb := make([]byte, 1500)
	for time.Now().Before(deadline) {
		n, _, readErr := conn.ReadFrom(rb)
		if readErr != nil {
			return 0, readErr
		}
		rm, parseErr := icmp.ParseMessage(ipv4.ICMPTypeEcho.Protocol(), rb[:n])
		if parseErr != nil {
			continue
		}
		if rm.Type == ipv4.ICMPTypeEchoReply {
			return time.Since(start), nil
		}
	}

	return 0, context.DeadlineExceeded
}
