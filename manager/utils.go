package manager

import (
	"bytes"
	"io"
)

func teeReader(src io.Reader) (io.Reader, io.Reader) {
	var dup bytes.Buffer
	tee := io.TeeReader(src, &dup)

	return tee, &dup
}

func sliceFilter[T any](src []T, f func(T) bool) []T {
	var res []T
	for _, v := range src {
		if f(v) {
			res = append(res, v)
		}
	}

	return res
}
