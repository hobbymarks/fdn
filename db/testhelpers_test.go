package db

import "testing"

func resetSharedDBOnCleanup(t *testing.T) {
	t.Helper()
	ResetSharedDB()
	
}
