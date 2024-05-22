package easipfs

import (
	"bytes"
	"io"
)

func teeIoReader(src io.Reader) (io.Reader, io.Reader) {
	var dup bytes.Buffer
	tee := io.TeeReader(src, &dup)

	return tee, &dup
}
