package version

import (
	"fmt"
	"strconv"
	"strings"
)

type SerialImageVersion struct {
	tag    string
	tagNum int
}

// parseSerialTag parses the serial tag, validates the format and returns the serial number
func parseSerialTag(tag string) (int, error) {
	imageTagParts := strings.Split(tag, "-")
	if imageTagParts[0] != "dev" || len(imageTagParts) != 3 {
		return 0, fmt.Errorf("invalid tag format for serial: %s", tag)
	}
	return strconv.Atoi(imageTagParts[1])
}

// NewSerialImageVersion creates SerialImageVersion
func NewSerialImageVersion(tag string) (*SerialImageVersion, error) {
	v := &SerialImageVersion{}

	var err error
	v.tag = tag
	v.tagNum, err = parseSerialTag(tag)
	if err != nil {
		return nil, err
	}

	return v, nil
}

// GetTag is a getter for getting tag members
func (v *SerialImageVersion) GetTag() string {
	return v.tag
}

// Compare compares with the specified tag
func (v *SerialImageVersion) Compare(anotherTag string) (int, error) {
	anotherTagNum, err := parseSerialTag(anotherTag)
	if err != nil {
		return 0, err
	}
	switch {
	case anotherTagNum < v.tagNum:
		return 1, nil
	case anotherTagNum > v.tagNum:
		return -1, nil
	default:
		return 0, nil
	}
}
