package version

import (
	"testing"
)

func TestProdImageVersion_Compare(t *testing.T) {
	tests := []struct {
		name       string
		currentTag string
		newTag     string
		result     int
	}{
		{
			"currentTag is equal to newTag",
			"v1.0.0",
			"v1.0.0",
			0,
		},
		{
			"currentTag is equal to newTag (ignore build metadata)",
			"v1.0.0",
			"v1.0.0+001",
			0,
		},
		{
			"currentTag is newer than newTag",
			"v1.0.0",
			"v0.1.0",
			1,
		},
		{
			"newTag is newer than currentTag (patch)",
			"v1.0.0",
			"v1.0.1",
			-1,
		},
		{
			"newTag is newer than currentTag (minor)",
			"v1.0.0",
			"v1.1.0",
			-1,
		},
		{
			"newTag is newer than currentTag (major)",
			"v1.0.0",
			"v2.0.0",
			-1,
		},
		{
			"newTag is newer than currentTag (pre-release)",
			"v1.0.0-alpha.1",
			"v1.0.0",
			-1,
		},
		{
			"newTag is newer than currentTag (pre-release)",
			"v1.0.0-alpha.1",
			"v1.0.0-beta.1",
			-1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ver, err := NewSemanticImageVersion(test.currentTag)
			if err != nil {
				t.Errorf("got unexpected error: %s", err.Error())
			}
			result, err := ver.Compare(test.newTag)
			if err != nil {
				t.Errorf("got unexpected error: %s", err.Error())
			}
			if result != test.result {
				t.Errorf("expected %d, got %d", test.result, result)
			}
		})
	}
}
