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

func IsDisallowedChar(r rune) bool {
	specialChars := "!#$%&'*+-.^_`|~"
	for _, c := range specialChars {
		if c == r {
			return false
		}
	}

	if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' {
		return false
	}
	return true
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

	key := strings.ToLower(strings.TrimRight(headerParts[0], ":"))
	for _, c := range key {
		if IsDisallowedChar(c) == true {
			return 0, false, fmt.Errorf("invalid header key: %s", key)
		}
	}
	value := strings.ToLower(strings.TrimSpace(headerParts[1]))
	if h[key] != "" {
		currentValue := h[key]
		h[key] = currentValue + ", " + value
		return byteCount, false, nil
	}
	h[key] = value
	return byteCount, false, nil
}
