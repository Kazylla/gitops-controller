package version

type TagFormat int

const (
	TagFormatSerial TagFormat = iota
	TagFormatSemantic
)

type ImageVersion interface {
	GetTag() string
	Compare(string) (int, error)
}

// NewImageVersion creates ImageVersion according to TagFormat
func NewImageVersion(tag string, tagFmt TagFormat) (ImageVersion, error) {
	var err error
	var imageVer ImageVersion

	switch {
	case tagFmt == TagFormatSerial:
		imageVer, err = NewSerialImageVersion(tag)
		if err != nil {
			return nil, err
		}
	case tagFmt == TagFormatSemantic:
		imageVer, err = NewSemanticImageVersion(tag)
		if err != nil {
			return nil, err
		}
	}

	return imageVer, nil
}
