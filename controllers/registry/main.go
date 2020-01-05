package registry

import (
	"github.com/go-logr/logr"
	"github.com/kazylla/gitops-controller/controllers/version"
)

type RegType int

const (
	RegECR RegType = iota
)

type Registry interface {
	GetTags(currentTag string) ([]version.ImageVersion, error)
}

type AWSCred struct {
	Profile string
}

type Config struct {
	Type RegType

	// Common
	Path      string
	TagFormat version.TagFormat
	Log       logr.Logger

	// RegType=ECR
	AWSCred AWSCred
}

// NewRegistry creates Registry according to RegType
func NewRegistry(c Config) Registry {
	switch c.Type {
	case RegECR:
		return &ECRRegistry{
			Config: c,
		}
	default:
		return &ECRRegistry{
			Config: c,
		}
	}
}
