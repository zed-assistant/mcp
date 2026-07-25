package filesystem

import (
	"errors"
	"io"
	"os"
)

var reverseReadChunkSize int64 = 64 * 1024

func ReadLinesFromEnd(path string, maxLines int, match func(line string) bool) ([]string, error) {
	if maxLines <= 0 {
		return []string{}, nil
	}

	f, err := os.Open(path) // nolint:gosec
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()

	info, err := f.Stat()
	if err != nil {
		return nil, err
	}

	matched := make([]string, 0, maxLines)
	pos := info.Size()
	var carry []byte
	isFirstSegment := true

	emit := func(line string) bool {
		if isFirstSegment {
			isFirstSegment = false
			if line == "" {
				return len(matched) < maxLines
			}
		}
		if match == nil || match(line) {
			matched = append(matched, line)
		}
		return len(matched) < maxLines
	}

	for pos > 0 && len(matched) < maxLines {
		chunkLen := min(reverseReadChunkSize, pos)
		start := pos - chunkLen

		buf := make([]byte, chunkLen+int64(len(carry)))
		if _, err := f.ReadAt(buf[:chunkLen], start); err != nil && !errors.Is(err, io.EOF) {
			return nil, err
		}
		copy(buf[chunkLen:], carry)

		segEnd := len(buf)
		keepGoing := true
		for i := len(buf) - 1; i >= 0 && keepGoing; i-- {
			if buf[i] == '\n' {
				keepGoing = emit(string(buf[i+1 : segEnd]))
				segEnd = i
			}
		}

		if !keepGoing {
			break
		}

		if start == 0 {
			emit(string(buf[:segEnd]))
			break
		}

		carry = append([]byte(nil), buf[:segEnd]...)
		pos = start
	}

	return matched, nil
}
