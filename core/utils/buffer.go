package utils

import (
	"bytes"
	"sync"
)

type SafeBuffer struct {
	Buffer bytes.Buffer
	mu     sync.Mutex
}

func (w *SafeBuffer) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.Buffer.Write(p)
}
func (w *SafeBuffer) Bytes() []byte {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.Buffer.Bytes()
}
func (w *SafeBuffer) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.Buffer.Reset()
}
