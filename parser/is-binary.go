package parser

import (
	"io"
	"strings"

	"github.com/edelbrocken/engine.io/types"
)

// Returns true if obj is a Buffer or a File.
func IsBinary(data interface{}) bool {
	switch data.(type) {
	case *types.StringBuffer: // false
	case *strings.Reader: // false
	case []byte:
		return true
	case io.Reader:
		return true
	}
	return false
}

func HasBinary(data interface{}) bool {
	switch o := data.(type) {
	case nil:
		return false
	case []interface{}:
		for _, v := range o {
			if HasBinary(v) {
				return true
			}
		}
	case map[string]interface{}:
		for _, v := range o {
			if HasBinary(v) {
				return true
			}
		}
	}
	return IsBinary(data)
}
