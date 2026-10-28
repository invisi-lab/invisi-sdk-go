package native

import "bytes"

// Bound diagnostic memory even when a local tool produces excessive output.
// Failed executions never expose this buffer to callers.
type boundedOutput struct {
	buffer   bytes.Buffer
	overflow bool
}

func (b *boundedOutput) Write(data []byte) (int, error) {
	size := len(data)
	remaining := 64<<10 - b.buffer.Len()
	if len(data) > remaining {
		b.overflow = true
		data = data[:remaining]
	}
	_, _ = b.buffer.Write(data)
	return size, nil
}

func (b *boundedOutput) Bytes() []byte { return b.buffer.Bytes() }
func (b *boundedOutput) Len() int      { return b.buffer.Len() }
