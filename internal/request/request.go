package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/UUest/httpfromtcp/internal/headers"
)

type ParserState int

const (
	StateInitialized ParserState = iota
	StateParsingHeaders
	StateParsingBody
	StateDone
)

type Request struct {
	RequestLine    RequestLine
	Headers        headers.Headers
	Body           []byte
	ParserState    ParserState
	bodyLengthRead int
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const crlf = "\r\n"
const bufferSize = 8
const contentLengthKey = "Content-Length"

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0
	req := Request{
		ParserState:    ParserState(StateInitialized),
		Headers:        headers.NewHeaders(),
		Body:           []byte{},
		bodyLengthRead: 0,
	}
	for req.ParserState != StateDone {
		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		bytesRead, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if req.ParserState != StateDone {
					return nil, fmt.Errorf("incomplete request, in state: %d, read n bytes on EOF: %d", req.ParserState, bytesRead)
				}
				break
			}
			return nil, err
		}
		readToIndex += bytesRead

		bytesParsed, err := req.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[bytesParsed:])
		readToIndex -= bytesParsed
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
	totalBytesParsed := 0
	for r.ParserState != StateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}
		totalBytesParsed += n
		if n == 0 {
			break
		}
	}
	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.ParserState {
	case StateInitialized:
		requestLine, bytesParsed, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if bytesParsed == 0 {
			return bytesParsed, err
		}
		r.RequestLine = *requestLine
		r.ParserState = StateParsingHeaders
		return bytesParsed, nil
	case StateParsingHeaders:
		bytesParsed, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if done {
			r.ParserState = StateParsingBody
		}
		return bytesParsed, nil
	case StateParsingBody:
		contentL := r.Headers.Get(contentLengthKey)
		if contentL == "" {
			r.ParserState = StateDone
			return 0, nil
		}
		contentLength, err := strconv.Atoi(contentL)
		if err != nil {
			return 0, err
		}
		r.Body = append(r.Body, data...)
		r.bodyLengthRead += len(data)
		if r.bodyLengthRead > contentLength {
			return 0, fmt.Errorf("invalid body length, expected: %v, got: %v", contentLength, len(r.Body))
		}
		if r.bodyLengthRead == contentLength {
			r.ParserState = StateDone
		}
		return len(data), nil

	case StateDone:
		return 0, fmt.Errorf("error: trying to read data in a done state")
	default:
		return 0, fmt.Errorf("invalid parser state")
	}
}
