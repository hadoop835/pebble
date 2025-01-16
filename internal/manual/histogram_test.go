// Copyright 2024 The LevelDB-Go and Pebble Authors. All rights reserved. Use
// of this source code is governed by a BSD-style license that can be found in
// the LICENSE file.

package manual

import (
	"fmt"
	"testing"

	"github.com/cockroachdb/crlib/crstrings"
	"github.com/cockroachdb/datadriven"
)

func TestHistBucketAndSubBucket(t *testing.T) {
	tests := []struct {
		size              uintptr
		bucket, subBucket int
	}{
		{size: 0b1000, bucket: 0, subBucket: 0},
		{size: 0b1001, bucket: 0, subBucket: 1},
		{size: 0b1111, bucket: 0, subBucket: 7},
		{size: 0b10000, bucket: 1, subBucket: 0},
		{size: 0b11111, bucket: 1, subBucket: 7},
		{size: 0b10111, bucket: 1, subBucket: 3},
		{size: 0b1011100, bucket: 3, subBucket: 3},
		{size: 0b1011111111, bucket: 6, subBucket: 3},
		{size: 0b11000000000000000000, bucket: 16, subBucket: 4},
		{size: 0b000, bucket: 0, subBucket: 0},
		{size: 0b001, bucket: 0, subBucket: 0},
		{size: 0b110, bucket: 0, subBucket: 0},
		{size: 1 << 31, bucket: 26, subBucket: 7},
	}
	for _, tc := range tests {
		bucket, subBucket := histBucketAndSubBucket(tc.size)
		if bucket != tc.bucket || subBucket != tc.subBucket {
			t.Errorf("size=0b%b: got bucket=%d, subBucket=%d; want bucket=%d, subBucket=%d",
				tc.size, bucket, subBucket, tc.bucket, tc.subBucket)
		}
	}
}

func TestHistogram(t *testing.T) {
	var h histogram
	datadriven.RunTest(t, "testdata/histogram", func(t *testing.T, d *datadriven.TestData) string {
		switch d.Cmd {
		case "alloc":
			for _, l := range crstrings.Lines(d.Input) {
				var size uintptr
				_, _ = fmt.Sscanf(l, "%d", &size)
				h.RecordAlloc(size)
			}

		case "free":
			for _, l := range crstrings.Lines(d.Input) {
				var size uintptr
				_, _ = fmt.Sscanf(l, "%d", &size)
				h.RecordAlloc(size)
			}

		default:
			d.Fatalf(t, "unknown command %q", d.Cmd)
		}
		return h.String()
	})
}
