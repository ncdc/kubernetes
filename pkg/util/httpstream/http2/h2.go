package http2

import (
	"net/http"
	"time"
)

type Connection struct {
}

func (c *Connection) CreateStream(headers http.Header) (Stream, error) {
}

func (c *Connection) Close() error {
}

func (c *Connection) CloseChan() <-chan bool {
}

func (c *Connection) SetIdleTimeout(timeout time.Duration) {
}

type Stream struct {
}

func (s *Stream) Read(p []byte) (int, error) {
}

func (s *Stream) Write(p []byte) (int, error) {
}

func (s *Stream) Close() error {
}

func (s *Stream) Reset() error {
}

func (s *Stream) Headers() http.Header {
}

func (s *Stream) Identifier() uint32 {
}
