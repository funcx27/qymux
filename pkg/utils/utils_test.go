package utils

import (
	"io"
	"testing"
)

type mockCloser struct {
	closeFunc func() error
}

func (m *mockCloser) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

func TestCloseAll(t *testing.T) {
	t.Run("close multiple closers", func(t *testing.T) {
		c1 := &mockCloser{}
		c2 := &mockCloser{}
		c3 := &mockCloser{}

		err := CloseAll(c1, c2, c3)
		if err != nil {
			t.Errorf("CloseAll() error = %v", err)
		}
	})

	t.Run("with nil closers", func(t *testing.T) {
		err := CloseAll(nil, nil)
		if err != nil {
			t.Errorf("CloseAll() with nils error = %v", err)
		}
	})

	t.Run("first error is returned", func(t *testing.T) {
		expectedErr := io.EOF
		c1 := &mockCloser{closeFunc: func() error { return expectedErr }}
		c2 := &mockCloser{closeFunc: func() error { return io.ErrClosedPipe }}
		c3 := &mockCloser{}

		err := CloseAll(c1, c2, c3)
		if err != expectedErr {
			t.Errorf("CloseAll() error = %v, want %v", err, expectedErr)
		}
	})

	t.Run("mixed nil and closers", func(t *testing.T) {
		c1 := &mockCloser{}
		err := CloseAll(nil, c1, nil)
		if err != nil {
			t.Errorf("CloseAll() mixed error = %v", err)
		}
	})
}
