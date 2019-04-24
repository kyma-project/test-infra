package printer

import (
	"encoding/json"
	"io"
)

// NewWithOutput returns new LgoPrinter instance with defined output
func NewWithOutput(stream *json.Decoder, ids []string, out io.Writer) *LogPrinter {
	lp := New(stream, ids)
	lp.out = out
	return lp
}
