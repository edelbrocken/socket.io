package parser

import (
	"github.com/edelbrocken/engine.io/events"
	"github.com/edelbrocken/engine.io/types"
)

// A socket.io Encoder instance
type Encoder interface {
	Encode(*Packet) []types.BufferInterface
}

// A socket.io Decoder instance
type Decoder interface {
	events.EventEmitter

	Add(interface{}) error
	Destroy()
}

type Parser interface {
	// A socket.io Encoder instance
	Encoder() Encoder

	// A socket.io Decoder instance
	Decoder() Decoder
}

type parser struct {
}

func (p *parser) Encoder() Encoder {
	return NewEncoder()
}

func (p *parser) Decoder() Decoder {
	return NewDecoder()
}

func NewParser() Parser {
	return &parser{}
}
