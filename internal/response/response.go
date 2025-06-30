package response

import (
	"fmt"
	"io"

	"github.com/UUest/httpfromtcp/internal/headers"
)

type StatusCode int

const (
	OK                  StatusCode = 200
	BadRequest          StatusCode = 400
	InternalServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, status StatusCode) error {
	switch status {
	case OK:
		_, err := w.Write([]byte("HTTP/1.1 200 OK\r\n"))
		return err
	case BadRequest:
		_, err := w.Write([]byte("HTTP/1.1 400 Bad Request\r\n"))
		return err
	case InternalServerError:
		_, err := w.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n"))
		return err
	default:
		_, err := w.Write([]byte(fmt.Sprintf("HTTP/1.1 %v \r\n", status)))
		return err
	}
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()
	headers["Content-Length"] = fmt.Sprintf("%d", contentLen)
	headers["Connection"] = "close"
	headers["Content-Type"] = "text/plain"
	return headers
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for key, value := range headers {
		_, err := w.Write([]byte(fmt.Sprintf("%s: %s\r\n", key, value)))
		if err != nil {
			return err
		}
	}
	w.Write([]byte("\r\n"))
	return nil
}
