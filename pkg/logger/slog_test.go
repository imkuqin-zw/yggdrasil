package logger

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"testing"
	"testing/slogtest"

	"github.com/stretchr/testify/require"
)

func TestSlogtest(t *testing.T) {
	var buff = bytes.NewBuffer(nil)

	lg := NewLogger(LvDebug, NewMemoryWriter(buff, true, nil))
	//global.writer =
	handler := NewSlogHandler(lg, nil)
	err := slogtest.TestHandler(
		handler,
		func() []map[string]any {
			// Parse the newline-delimted JSON in buff.
			var entries []map[string]any
			//dec := json.NewDecoder(buff)
			reader := bufio.NewReader(buff)
			for {
				line, _, err := reader.ReadLine()
				if err != nil {
					if err == io.EOF {
						break
					}
					continue
				}
				var ent map[string]any
				require.NoError(t, json.Unmarshal(line, &ent), "Error decoding log message")
				entries = append(entries, ent)
			}
			return entries
		},
	)
	require.NoError(t, err, "Unexpected error from slogtest.TestHandler")
}
