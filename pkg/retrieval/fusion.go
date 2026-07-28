package retrieval

// RRF combines ranked lists with reciprocal rank fusion.
func RRF(lists [][]RankedHit, k int, weights []float64) []RankedHit {
	if k <= 0 {
		k = DefaultRRFK
	}
	if len(weights) == 0 {
		weights = make([]float64, len(lists))
		for i := range weights {
			weights[i] = 1.0
		}
	}
	scores := map[string]float64{}
	hits := map[string]RankedHit{}
	for li, list := range lists {
		w := 1.0
		if li < len(weights) {
			w = weights[li]
		}
		for rank, hit := range list {
			sid := hit.SkillID
			if sid == "" {
				continue
			}
			scores[sid] += w * (1.0 / float64(k+rank+1))
			if prev, ok := hits[sid]; !ok || hit.Score > prev.Score {
				h := hit
				h.Score = scores[sid]
				h.Hit.MatchType = MatchHybrid
				hits[sid] = h
			}
		}
	}
	out := make([]RankedHit, 0, len(hits))
	for sid, score := range scores {
		h := hits[sid]
		h.Score = score
		out = append(out, h)
	}
	sortRanked(out)
	return out
}

func sortRanked(h []RankedHit) {
	for i := 0; i < len(h); i++ {
		for j := i + 1; j < len(h); j++ {
			if h[j].Score > h[i].Score {
				h[i], h[j] = h[j], h[i]
			}
		}
	}
}
