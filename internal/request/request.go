package request

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

type ParserState int

const (
	StateInitialized ParserState = iota
	StateDone        ParserState = 1
)

type Request struct {
	RequestLine RequestLine
	ParserState ParserState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const crlf = "\r\n"
const bufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0
	req := Request{
		ParserState: ParserState(StateInitialized),
	}
	for req.ParserState != StateDone {
		if len(buf) == cap(buf) {
			newBuf := make([]byte, cap(buf)*2)
			copy(newBuf, buf[:readToIndex])
			buf = newBuf
		}
		bytesRead, err := reader.Read(buf[readToIndex:])
		var bytesParsed int
		if bytesRead > 0 {
			readToIndex += bytesRead
			bytesParsed, err = req.parse(buf[:readToIndex])
			copy(buf, buf[bytesParsed:readToIndex])
			readToIndex -= bytesParsed
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}
	return &req, nil
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return nil, 0, nil
	}
	byteCount := idx + len(crlf)
	requestLineText := string(data[:idx])
	requestLine, err := requestLineFromString(requestLineText)
	if err != nil {
		return nil, 0, err
	}
	return requestLine, byteCount, nil
}

func requestLineFromString(str string) (*RequestLine, error) {
	parts := strings.Split(str, " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("poorly formatted request-line: %s", str)
	}

	method := parts[0]
	for _, c := range method {
		if c < 'A' || c > 'Z' {
			return nil, fmt.Errorf("invalid method: %s", method)
		}
	}

	requestTarget := parts[1]

	versionParts := strings.Split(parts[2], "/")
	if len(versionParts) != 2 {
		return nil, fmt.Errorf("malformed start-line: %s", str)
	}

	httpPart := versionParts[0]
	if httpPart != "HTTP" {
		return nil, fmt.Errorf("unrecognized HTTP-version: %s", httpPart)
	}
	version := versionParts[1]
	if version != "1.1" {
		return nil, fmt.Errorf("unrecognized HTTP-version: %s", version)
	}

	return &RequestLine{
		Method:        method,
		RequestTarget: requestTarget,
		HttpVersion:   versionParts[1],
	}, nil
}
func (r *Request) parse(data []byte) (int, error) {
	if r.ParserState == StateInitialized {
		requestLine, bytesParsed, err := parseRequestLine(data)
		if err != nil {
			return bytesParsed, err
		}
		if bytesParsed == 0 {
			return bytesParsed, err
		}
		if bytesParsed > 0 {
			r.RequestLine = *requestLine
			r.ParserState = StateDone
			return bytesParsed, nil
		}
	} else if r.ParserState == StateDone {
		return 0, fmt.Errorf("error: trying to read data in a done state")
	}
	return 0, fmt.Errorf("invalid parser state")
}
