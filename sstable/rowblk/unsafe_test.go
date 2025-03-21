// Copyright 2018 The LevelDB-Go and Pebble Authors. All rights reserved. Use
// of this source code is governed by a BSD-style license that can be found in
// the LICENSE file.

package rowblk

import (
	"encoding/binary"
	"fmt"
	"io"
	"math/rand/v2"
	"testing"
	"time"
	"unsafe"

	"github.com/cockroachdb/crlib/testutils/leaktest"
)

func TestDecodeVarint(t *testing.T) {
	defer leaktest.AfterTest(t)()
	vals := []uint32{
		0,
		1,
		1 << 7,
		1 << 8,
		1 << 14,
		1 << 15,
		1 << 20,
		1 << 21,
		1 << 28,
		1 << 29,
		1 << 31,
	}
	buf := make([]byte, 5)
	for _, v := range vals {
		binary.PutUvarint(buf, uint64(v))
		u, _ := decodeVarint(unsafe.Pointer(&buf[0]))
		if v != u {
			fmt.Printf("%d %d\n", v, u)
		}
	}
}

func BenchmarkDecodeVarint(b *testing.B) {
	rng := rand.New(rand.NewPCG(0, uint64(time.Now().UnixNano())))
	vals := make([]unsafe.Pointer, 10000)
	for i := range vals {
		buf := make([]byte, 5)
		binary.PutUvarint(buf, uint64(rng.Uint32()))
		vals[i] = unsafe.Pointer(&buf[0])
	}

	b.ResetTimer()
	var ptr unsafe.Pointer
	for i, n := 0, 0; i < b.N; i += n {
		n = len(vals)
		if n > b.N-i {
			n = b.N - i
		}
		for j := 0; j < n; j++ {
			_, ptr = decodeVarint(vals[j])
		}
	}
	if testing.Verbose() {
		fmt.Fprint(io.Discard, ptr)
	}
}
