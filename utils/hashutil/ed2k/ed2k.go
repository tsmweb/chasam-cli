// Package ed2k implements the ed2k hash algorithm.
package ed2k

import (
	"golang.org/x/crypto/md4"
	"hash"
)

const (
	chunk = 9728000
)

type digest struct {
	inner hash.Hash
	total hash.Hash
	round int
	s     int
}

// New returns a new hash.Hash computing the ed2k checksum.
func New() hash.Hash {
	d := new(digest)
	d.inner = md4.New()
	d.total = md4.New()
	return d
}

func (d *digest) Write(b []byte) (nn int, err error) {
	nn = len(b)
	if nn+d.s < chunk {
		d.inner.Write(b)
		d.s += nn
		return
	}

	l := chunk - d.s
	d.inner.Write(b[:l])
	d.total.Write(d.inner.Sum(nil))
	d.next()
	left := nn - l
	min := chunk

	for left > 0 {
		if left < chunk {
			min = left
		}

		d.inner.Write(b[l : l+min])

		if min == chunk {
			d.total.Write(d.inner.Sum(nil))
			d.next()
		} else {
			d.s = min
		}

		left -= min
	}

	return
}

func (d *digest) next() {
	d.round++
	d.s = 0
	d.inner.Reset()
}

func (d *digest) Sum(in []byte) []byte {
	dd := *d
	if dd.round == 0 {
		return dd.inner.Sum(in)
	}
	if dd.s > 0 {
		dd.total.Write(dd.inner.Sum(nil))
	}
	return dd.total.Sum(in)
}

func (d *digest) Reset() {
	d.inner.Reset()
	d.total.Reset()
	d.round = 0
	d.s = 0
}

func (d *digest) Size() int {
	return 16
}

func (d *digest) BlockSize() int {
	return chunk
}
