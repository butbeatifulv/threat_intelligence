package feeds

import (
	"net"
	"net/url"
	"strings"

	"github.com/butbeautifulv/veil/pkg/ti/domain"
)

// IOCFromThreatFoxExport maps ThreatFox export / API row fields to a domain IOC (normalize in pipeline).
func IOCFromThreatFoxExport(iocValue, iocType string) (domain.IOC, bool) {
	t := strings.ToLower(strings.TrimSpace(iocType))
	v := strings.TrimSpace(iocValue)
	if v == "" {
		return domain.IOC{}, false
	}
	base := domain.IOC{Source: "threatfox"}
	switch t {
	case "domain":
		base.Type = domain.IOCDomain
		base.Value = v
	case "url":
		base.Type = domain.IOCURL
		base.Value = v
	case "ip:port":
		host, _, err := net.SplitHostPort(v)
		if err != nil {
			if i := strings.LastIndex(v, ":"); i > 0 && net.ParseIP(v[:i]) != nil {
				host = v[:i]
			} else {
				host = v
			}
		}
		base.Type = domain.IOCIP
		base.Value = host
	case "ipv4", "ip":
		base.Type = domain.IOCIP
		base.Value = v
	case "ipv6":
		base.Type = domain.IOCIP
		base.Value = v
	case "md5_hash", "md5":
		base.Type = domain.IOCHash
		base.Value = v
	case "sha256_hash", "sha256":
		base.Type = domain.IOCHash
		base.Value = v
	case "sha1_hash", "sha1":
		base.Type = domain.IOCHash
		base.Value = v
	default:
		vs := strings.ToLower(v)
		if strings.HasPrefix(vs, "http://") || strings.HasPrefix(vs, "https://") {
			base.Type = domain.IOCURL
			base.Value = v
			break
		}
		if ip := net.ParseIP(v); ip != nil {
			base.Type = domain.IOCIP
			base.Value = v
			break
		}
		if u, err := url.Parse(v); err == nil && u.Scheme != "" && u.Host != "" {
			base.Type = domain.IOCURL
			base.Value = v
			break
		}
		if !strings.ContainsAny(v, "/:") {
			base.Type = domain.IOCDomain
			base.Value = v
			break
		}
		return domain.IOC{}, false
	}
	return base, true
}
