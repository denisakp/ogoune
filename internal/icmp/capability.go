package icmp

import (
	"context"
	"net"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// CapabilityResult reports whether this host can send ICMP echo probes.
type CapabilityResult struct {
	Available bool
	Reason    string
}

// Detect tests whether the current process can open an ICMP raw/privileged socket.
// It attempts to open both the privileged raw socket and the unprivileged echo socket
// (Linux ping_group_range / macOS) and returns Available=true if either succeeds.
// Errors are reported in Reason without panicking.
func Detect() CapabilityResult {
	// Try unprivileged datagram ICMP socket first (works on Linux with
	// net.ipv4.ping_group_range and on macOS by default).
	conn, err := icmp.ListenPacket("udp4", "0.0.0.0")
	if err == nil {
		conn.Close()
		return CapabilityResult{Available: true}
	}

	// Fall back to privileged raw socket (requires root or CAP_NET_RAW).
	conn, err = icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err == nil {
		conn.Close()
		return CapabilityResult{Available: true}
	}

	return CapabilityResult{
		Available: false,
		Reason:    "cannot open ICMP socket: requires root, CAP_NET_RAW, or net.ipv4.ping_group_range; " + err.Error(),
	}
}

// probeICMPEcho sends a single ICMP echo request to addr and returns round-trip time.
// It uses unprivileged UDP ping where available, falling back to raw IP.
// timeout controls the deadline for the round-trip.
func probeICMPEcho(addr string, timeout time.Duration) (rtt time.Duration, err error) {
	// Resolve target first
	ips, err := net.LookupHost(addr)
	if err != nil || len(ips) == 0 {
		return 0, err
	}
	target := ips[0]

	// Try unprivileged first
	conn, dialErr := icmp.ListenPacket("udp4", "0.0.0.0")
	network := "udp4"
	if dialErr != nil {
		conn, dialErr = icmp.ListenPacket("ip4:icmp", "0.0.0.0")
		network = "ip4:icmp"
		if dialErr != nil {
			return 0, dialErr
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

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	start := time.Now()

	var dst net.Addr
	if network == "udp4" {
		dst = &net.UDPAddr{IP: net.ParseIP(target)}
	} else {
		dst = &net.IPAddr{IP: net.ParseIP(target)}
	}

	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return 0, err
	}

	if _, err := conn.WriteTo(wb, dst); err != nil {
		return 0, err
	}

	rb := make([]byte, 1500)
	for {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
		}
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
}
