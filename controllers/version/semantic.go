package version

import (
	"fmt"
	"strings"

	version2 "k8s.io/apimachinery/pkg/util/version"
)

type SemanticImageVersion struct {
	tag    string
	semVer *version2.Version
}

// NewSemanticImageVersion creates SemanticImageVersion
func NewSemanticImageVersion(tag string) (*SemanticImageVersion, error) {
	v := &SemanticImageVersion{}

	if !strings.HasPrefix(tag, "v") {
		return nil, fmt.Errorf("invalid tag format for semantic: %s", tag)
	}

	var err error
	v.tag = tag
	v.semVer, err = version2.ParseSemantic(tag)
	if err != nil {
		return nil, err
	}

	return v, nil
}

// GetTag is a getter for getting tag members
func (v *SemanticImageVersion) GetTag() string {
	return v.tag
}

// Compare compares with the specified tag
func (v *SemanticImageVersion) Compare(anotherTag string) (int, error) {
	return v.semVer.Compare(anotherTag)
}
