// Copyright 2024 The LevelDB-Go and Pebble Authors. All rights reserved. Use
// of this source code is governed by a BSD-style license that can be found in
// the LICENSE file.

package manual

import (
	"bytes"
	"fmt"
	"math/bits"
	"sync/atomic"
	"text/tabwriter"

	"github.com/cockroachdb/pebble/internal/humanize"
)

// histogram of allocations by size.
//
// The histogram is organized in buckets that grow exponentially (each bucket is
// 2x wider than the previous one), and each bucket is divided evenly into 8
// sub-buckets.
//
// We keep track of total allocs and total frees, allowing to see both the
// current usage and the cumulative number of allocations in each subbucket.
type histogram struct {
	allocs histCounters
	frees  histCounters
}

const histMaxBits = 30
const histSubBucketBits = 3
const histSubBuckets = 1 << histSubBucketBits

type histCounters [histMaxBits - histSubBucketBits][histSubBuckets]atomic.Uint32

// histBucketAndSubBucket determines the bucket and sub bucket for a given allocation size.
//
// The position of the highest bit determines the bucket, and the values of the
// following 3 bits determine the sub-bucket:
//
//	   bucket
//	   v
//	00010110101001111
//	    ^^^
//	    sub-bucket
func histBucketAndSubBucket(size uintptr) (bucket, subBucket int) {
	size = min(size, (1<<histMaxBits)-1)
	highBit := 31 - bits.LeadingZeros32(uint32(size))
	if highBit < histSubBucketBits {
		return 0, 0
	}
	bucket = highBit - histSubBucketBits
	subBucket = int((size >> (highBit - histSubBucketBits)) & (histSubBuckets - 1))
	return bucket, subBucket
}

func (h *histogram) RecordAlloc(size uintptr) {
	bucket, subBucket := histBucketAndSubBucket(size)
	h.allocs[bucket][subBucket].Add(1)
}

func (h *histogram) RecordFree(size uintptr) {
	bucket, subBucket := histBucketAndSubBucket(size)
	h.frees[bucket][subBucket].Add(1)
}

func (h *histogram) String() string {
	var buf bytes.Buffer
	tw := tabwriter.NewWriter(&buf, 2, 1, 4, ' ', 0)
	_, _ = fmt.Fprintf(tw, "start\twidth\tin use\ttotal\n")
	for b := 0; b < histMaxBits-histSubBucketBits; b++ {
		var allocs, frees [histSubBuckets]uint32
		for j := range allocs {
			allocs[j] = h.allocs[b][j].Load()
			frees[j] = h.frees[b][j].Load()
		}
		bucketLow := 1 << (histSubBucketBits + b)
		subBucketWidth := 1 << b
		for j, a := range allocs {
			if a == 0 {
				continue
			}
			start := humanize.Bytes.Int64(int64(bucketLow + subBucketWidth*j))
			width := humanize.Bytes.Int64(int64(subBucketWidth))
			if b == 0 && j == 0 {
				start = "0"
				width = humanize.Bytes.Int64(int64(bucketLow + subBucketWidth))
			} else if b == histMaxBits-histSubBucketBits-1 && j == histSubBuckets-1 {
				width = ""
			}
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
				start, width,
				humanize.Count.Int64(int64(a-frees[j])),
				humanize.Count.Int64(int64(a)),
			)
		}
	}
	_ = tw.Flush()
	return buf.String()
}
