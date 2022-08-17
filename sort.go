package main

import "google.golang.org/api/youtube/v3"

func sort(items []*youtube.PlaylistItem) {
	// Find longest monotonic subsequence
	ss := longestMonotonicSubseq(items)
	_ = ss
}

func longestMonotonicSubseq(items []*youtube.PlaylistItem) []bool {
	type subseq struct {
		seq  []bool
		len  int // number of `true` values in `seq`
		last *string
	}

	// Keep a list of good candidates so far
	// First possible candidate is empty subseq
	candidates := []subseq{{[]bool{}, 0, nil}}

	for _, it := range items {
		newCandidates := []subseq{}
		// Keep track of candidates to be discarded
		toDiscard := make([]bool, len(candidates))

		for _, oldSeq := range candidates {
			if oldSeq.last == nil || *oldSeq.last < it.Snippet.Title {
				// seq with next item added is a new candidate
				newSeq := subseq{
					seq:  append(oldSeq.seq, true),
					len:  oldSeq.len + 1,
					last: &it.Snippet.Title,
				}
				newCandidates = append(newCandidates, newSeq)

				// Check if this sequence "beats" any existing candidate
				for i, cd := range candidates {
					if cd.last != nil && cd.len <= newSeq.len && *cd.last >= *newSeq.last {
						// We could replace `cd` with `newSeq` inside any sequence to get
						// a better candidate. Therefore `cd` is not a candidate anymore.
						toDiscard[i] = true
					}
				}
			}
		}

		// Carry over existing candidates (which aren't marked to be discarded)
		for i, oldSeq := range candidates {
			if !toDiscard[i] {
				newCandidates = append(candidates, subseq{
					seq:  append(oldSeq.seq, false),
					len:  oldSeq.len,
					last: oldSeq.last,
				})
			}
		}

		candidates = newCandidates
	}

	// Find best candidate at end
	best := candidates[0]
	for _, cd := range candidates[1:] {
		if cd.len > best.len {
			best = cd
		}
	}

	return best.seq
}
