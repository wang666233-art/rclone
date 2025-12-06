package baidupan

import (
	"testing"

	"github.com/rclone/rclone/fstest/fstests"
)

// TestIntegration runs integration tests
func TestIntegration(t *testing.T) {
	fstests.Run(t, &fstests.Opt{
		RemoteName: "TestBaiduPan:",
		NilObject:  (*Object)(nil),
	})
}








<<<<<<< HEAD

=======
>>>>>>> 924501008797cef4628c9b23822429374e50c84d
