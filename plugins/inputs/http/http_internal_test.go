package http

import (
	"io"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

type closeTrackingReadCloser struct {
	closed *atomic.Int32
}

func (*closeTrackingReadCloser) Read(_ []byte) (int, error) {
	return 0, io.EOF
}

func (c *closeTrackingReadCloser) Close() error {
	c.closed.Add(1)
	return nil
}

func TestGatherURLClosesBodyOnRequestBuildError(t *testing.T) {
	originalFactory := requestBodyReaderFactory
	t.Cleanup(func() {
		requestBodyReaderFactory = originalFactory
	})

	var closeCalls atomic.Int32
	requestBodyReaderFactory = func(_, _ string) io.Reader {
		return &closeTrackingReadCloser{closed: &closeCalls}
	}

	h := &HTTP{
		Method: "BAD METHOD",
		Body:   "payload",
	}

	err := h.gatherURL(nil, "http://example.com")
	require.Error(t, err)
	require.Equal(t, int32(1), closeCalls.Load())
}

func TestGatherURLClosesBodyOnEarlyReturnAfterRequest(t *testing.T) {
	originalFactory := requestBodyReaderFactory
	t.Cleanup(func() {
		requestBodyReaderFactory = originalFactory
	})

	var closeCalls atomic.Int32
	requestBodyReaderFactory = func(_, _ string) io.Reader {
		return &closeTrackingReadCloser{closed: &closeCalls}
	}

	h := &HTTP{
		Method:    "GET",
		Body:      "payload",
		TokenFile: "does-not-exist.token",
	}

	err := h.gatherURL(nil, "http://example.com")
	require.Error(t, err)
	require.Equal(t, int32(1), closeCalls.Load())
}
