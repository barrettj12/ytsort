package main

import (
	"google.golang.org/api/youtube/v3"
)

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

	n := len(items)

	// For each i = 0, ..., n, we keep track of the longest monotonic subsequence
	lms := make([]*subseq, n+1)
	lms[0] = &subseq{
		seq:  make([]bool, n),
		len:  0,
		last: nil,
	}

	// Also keep track of sequences which have been "dominated"
	// S dominates T if len(S) >= len(T) and last(S) <= last(T)
	dominated := make([]bool, n+1)

	for i, it := range items {
		// Find new candidate lms[i+1]
		// which is the longest lms[j], j < i such that `it` can be added to the end
		j := 0
		for k := 1; k <= i; k++ {
			if !dominated[k] && *lms[k].last < it.Snippet.Title && lms[k].len > lms[j].len {
				j = k
			}
		}

		// Add new candidate to lms array
		newSeq := make([]bool, n)
		copy(newSeq, lms[j].seq)
		newSeq[i] = true
		lms[i+1] = &subseq{
			seq:  newSeq,
			len:  lms[j].len + 1,
			last: &it.Snippet.Title,
		}

		// Check which sequences are dominated by the new candidate
		// and mark these as such.
		for k := 1; k <= i; k++ {
			if !dominated[k] && *lms[k].last >= *lms[i+1].last && lms[k].len <= lms[i+1].len {
				dominated[k] = true
			}
		}
	}

	// Go through and pick the best sequence
	best := lms[0]
	for _, ss := range lms[1:] {
		if ss.len > best.len {
			best = ss
		}
	}
	return best.seq
}
