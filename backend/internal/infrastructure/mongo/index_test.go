package mongo

import "testing"

func TestLegacyIndexSpecs(t *testing.T) {
	legacy := legacyIndexSpecs()
	if len(legacy) != 1 {
		t.Fatalf("legacy index count = %d, want 1", len(legacy))
	}
	if legacy[0].collection != CollContracts {
		t.Fatalf("legacy collection = %s, want %s", legacy[0].collection, CollContracts)
	}
	if legacy[0].name != "ix_taskId" {
		t.Fatalf("legacy name = %s, want ix_taskId", legacy[0].name)
	}
}
