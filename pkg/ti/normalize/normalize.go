// Package normalize provides TI entity normalization for pipeline NED (SOT).
// Graph ingest must not import this package — payloads arrive pre-normalized via commit envelopes.
package normalize

import (
	"net"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"github.com/butbeautifulv/veil/pkg/ti/domain"
	"github.com/butbeautifulv/veil/pkg/ti/ids"
	"github.com/butbeautifulv/veil/pkg/ti/validate"
)

var (
	spaceRe   = regexp.MustCompile(`\s+`)
	hexHashRe = regexp.MustCompile(`^[a-f0-9]{32}$|^[a-f0-9]{40}$|^[a-f0-9]{64}$`)
)

// CanonicalID returns the canonical Neo4j IOC node id (alias of ids.CanonicalIOCID).
func CanonicalID(i domain.IOC) string {
	return ids.CanonicalIOCID(i)
}

// ActorStableID forwards to pkg/ti/ids.
func ActorStableID(name string) string { return ids.ActorStableID(name) }

// ReportStableID forwards to pkg/ti/ids.
func ReportStableID(link string) string { return ids.ReportStableID(link) }

// NormalizeIOC normalizes and validates an IOC; false if invalid after normalization rules.
func NormalizeIOC(i domain.IOC) (domain.IOC, bool) {
	if err := validate.CheckIOCShape(i); err != nil {
		return domain.IOC{}, false
	}
	return normalizeValidatedIOC(i)
}

func normalizeValidatedIOC(i domain.IOC) (domain.IOC, bool) {
	i.Value = strings.TrimSpace(i.Value)
	i.Source = strings.TrimSpace(i.Source)
	var srcs []string
	seen := map[string]struct{}{}
	add := func(s string) {
		s = strings.TrimSpace(s)
		if s == "" {
			return
		}
		if _, ok := seen[s]; ok {
			return
		}
		seen[s] = struct{}{}
		srcs = append(srcs, s)
	}
	for _, s := range i.Sources {
		add(s)
	}
	if i.Source != "" {
		add(i.Source)
	}
	sort.Strings(srcs)
	i.Sources = srcs

	switch i.Type {
	case domain.IOCIP:
		ip := net.ParseIP(i.Value)
		if ip == nil {
			return domain.IOC{}, false
		}
		i.Value = ip.String()
		return i, true
	case domain.IOCDomain:
		v := strings.ToLower(strings.TrimSuffix(i.Value, "."))
		v = spaceRe.ReplaceAllString(v, "")
		if v == "" || strings.ContainsAny(v, "/:@") {
			return domain.IOC{}, false
		}
		i.Value = v
		return i, true
	case domain.IOCURL:
		u, err := url.Parse(strings.TrimSpace(i.Value))
		if err != nil || u.Scheme == "" || u.Host == "" {
			return domain.IOC{}, false
		}
		u.Scheme = strings.ToLower(u.Scheme)
		u.Host = strings.ToLower(u.Host)
		if (u.Scheme == "http" && strings.HasSuffix(u.Host, ":80")) || (u.Scheme == "https" && strings.HasSuffix(u.Host, ":443")) {
			u.Host = strings.Split(u.Host, ":")[0]
		}
		i.Value = u.String()
		return i, true
	case domain.IOCHash:
		v := strings.ToLower(spaceRe.ReplaceAllString(i.Value, ""))
		if !hexHashRe.MatchString(v) {
			return domain.IOC{}, false
		}
		i.Value = v
		return i, true
	default:
		return domain.IOC{}, false
	}
}

// NormalizeCampaign trims campaign fields.
func NormalizeCampaign(c domain.Campaign) domain.Campaign {
	c.ID = strings.TrimSpace(c.ID)
	c.Name = strings.TrimSpace(c.Name)
	c.Summary = strings.TrimSpace(c.Summary)
	c.Source = strings.TrimSpace(c.Source)
	for i := range c.Actors {
		c.Actors[i] = strings.TrimSpace(c.Actors[i])
	}
	return c
}

// NormalizeCluster trims cluster fields.
func NormalizeCluster(cl domain.Cluster) domain.Cluster {
	cl.ID = strings.TrimSpace(cl.ID)
	cl.Name = strings.TrimSpace(cl.Name)
	cl.Description = strings.TrimSpace(cl.Description)
	cl.Source = strings.TrimSpace(cl.Source)
	return cl
}
