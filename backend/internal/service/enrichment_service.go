package service

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/likexian/whois"
	whoisparser "github.com/likexian/whois-parser"
)

type SSLInfo struct {
	ExpirationDate *time.Time
	Issuer         string
}

type WhoisInfo struct {
	ExpirationDate *time.Time
	Registrar      string
}

type EnrichmentService struct {
	defaultTimeout time.Duration
}

func NewEnrichmentService(timeout time.Duration) *EnrichmentService {
	return &EnrichmentService{
		defaultTimeout: timeout,
	}
}

// Enrich collect metadata from the resource Target
func (s *EnrichmentService) Enrich(ctx context.Context, resource *domain.Resource) (*domain.ResourceMetaData, error) {
	metadata := &domain.ResourceMetaData{}

	hostname := s.extractHostname(resource.Target)
	if hostname == "" {
		return metadata, nil
	}

	if resource.Type == domain.ResourceHTTP {
		if sslInfo := s.collectSSLInfo(ctx, hostname, resource.Timeout); sslInfo != nil {
			metadata.SSLExpirationDate = sslInfo.ExpirationDate
			metadata.SSLIssuer = sslInfo.Issuer
		}

		if whoisInfo := s.collectWhoisInfo(ctx, hostname); whoisInfo != nil {
			metadata.DomainExpirationDate = whoisInfo.ExpirationDate
			metadata.DomainRegistrar = whoisInfo.Registrar
		}
	}

	return metadata, nil
}

// extractHostname extract hostname from url or return the target as it is
func (s *EnrichmentService) extractHostname(target string) string {
	if u, err := url.Parse(target); err == nil && u.Host != "" {
		host := u.Host

		if h, _, err := net.SplitHostPort(host); err == nil {
			return h
		}
		return host
	}

	if host, _, err := net.SplitHostPort(target); err == nil {
		return host
	}

	return target
}

func (s *EnrichmentService) collectSSLInfo(ctx context.Context, hostname string, timeout int) *SSLInfo {
	if timeout <= 0 {
		timeout = int(10 * time.Second)
	}

	dialer := &net.Dialer{
		Timeout: time.Duration(timeout) * time.Second,
	}

	conn, err := tls.DialWithDialer(
		dialer,
		"tcp",
		fmt.Sprintf("%s:443", hostname),
		&tls.Config{
			InsecureSkipVerify: true,
			ServerName:         hostname,
		},
	)

	if err != nil {
		return nil // ssl not available and this is okay
	}
	defer conn.Close()

	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return nil
	}

	cert := certs[0]
	return &SSLInfo{
		ExpirationDate: &cert.NotAfter,
		Issuer: func() string {
			if len(cert.Issuer.Organization) > 0 {
				return cert.Issuer.Organization[0]
			}
			return cert.Issuer.CommonName
		}(), // this show the issuer instead of the intermediary certificate
	}
}

func (s *EnrichmentService) collectWhoisInfo(ctx context.Context, hostname string) *WhoisInfo {
	raw, err := whois.Whois(hostname)
	if err != nil {
		return nil
	}

	info, err := whoisparser.Parse(raw)
	if err != nil {
		return nil
	}

	expStr := info.Domain.ExpirationDate
	if expStr == "" {
		return nil
	}

	exp, err := time.Parse("2006-01-02T15:04:05Z", expStr)
	if err != nil {
		exp, err = time.Parse("2006-01-02", expStr)
		if err != nil {
			return nil
		}
	}

	return &WhoisInfo{
		ExpirationDate: &exp,
		Registrar:      info.Registrar.Name,
	}
}
