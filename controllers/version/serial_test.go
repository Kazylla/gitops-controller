package version

import (
	"testing"
)

func TestDevImageVersion_Compare(t *testing.T) {
	tests := []struct {
		name       string
		currentTag string
		newTag     string
		result     int
	}{
		{
			"currentTag is equal to newTag",
			"dev-100-xxxxxx",
			"dev-100-xxxxxx",
			0,
		},
		{
			"currentTag is newer than newTag",
			"dev-101-xxxxxx",
			"dev-100-xxxxxx",
			1,
		},
		{
			"newTag is newer than currentTag",
			"dev-100-xxxxxx",
			"dev-101-xxxxxx",
			-1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ver, err := NewSerialImageVersion(test.currentTag)
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
