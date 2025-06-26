package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

const crlf = "\r\n"

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return 0, false, nil
	}
	if idx == 0 {
		return idx + len(crlf), true, nil
	}
	byteCount := idx + len(crlf)
	headerText := string(data[:idx])
	header := strings.TrimSpace(headerText)
	headerParts := strings.SplitN(header, " ", 2)
	keyLength := len(headerParts[0])
	if headerParts[0][keyLength-1] != ':' {
		return 0, false, fmt.Errorf("invalid header: %s", headerText)
	}
	if len(headerParts) != 2 {
		return 0, false, fmt.Errorf("invalid header: %s", headerText)
	}
	h[strings.TrimRight(headerParts[0], ":")] = strings.TrimSpace(headerParts[1])
	return byteCount, false, nil
}
