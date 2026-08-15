package teldrive

import (
	"testing"

	"github.com/rclone/rclone/fstest/fstests"
)

func TestJoinAPIURLPreservesPathPrefix(t *testing.T) {
	for _, test := range []struct {
		base, endpoint, want string
	}{
		{"https://example.com/tgd", "/api/v1/me", "https://example.com/tgd/api/v1/me"},
		{"https://example.com/tgd/", "/api/v1/files", "https://example.com/tgd/api/v1/files"},
		{"http://127.0.0.1:54807", "/api/v1/me", "http://127.0.0.1:54807/api/v1/me"},
	} {
		if got := joinAPIURL(test.base, test.endpoint); got != test.want {
			t.Fatalf("joinAPIURL(%q, %q) = %q, want %q", test.base, test.endpoint, got, test.want)
		}
	}
}

// TestIntegration runs integration tests against the remote
func TestIntegration(t *testing.T) {
	fstests.Run(t, &fstests.Opt{
		RemoteName: "TestTeldrive:",
		NilObject:  (*Object)(nil),
		ChunkedUpload: fstests.ChunkedUploadConfig{
			MinChunkSize:  minChunkSize,
			CeilChunkSize: fstests.NextPowerOfTwo,
		},
		SkipInvalidUTF8: true,
	})
}
