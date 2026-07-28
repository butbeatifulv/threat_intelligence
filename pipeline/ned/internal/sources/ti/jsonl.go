package ti

import tidomain "github.com/butbeautifulv/veil/pkg/ti/domain"

// JSONLEnvelope is a single TI JSONL record (ioc, campaign, cluster, actor, or report).
type JSONLEnvelope struct {
	IOC      *tidomain.IOC      `json:"ioc,omitempty"`
	Campaign *tidomain.Campaign `json:"campaign,omitempty"`
	Cluster  *tidomain.Cluster  `json:"cluster,omitempty"`
	Actor    *tidomain.Actor    `json:"actor,omitempty"`
	Report   *tidomain.Report   `json:"report,omitempty"`
}
