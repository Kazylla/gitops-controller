package registry

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/kazylla/gitops-controller/controllers/version"
)

type ECRRegistry struct {
	Config Config
}

type ECRRegistryPath struct {
	AWSAccountID string
	Region       string
	Repo         string
}

// parseRegistryPath parses ECR registry path
func (e *ECRRegistry) parseRegistryPath() (*ECRRegistryPath, error) {
	// validate ecr registry path format
	pathParts := strings.SplitN(e.Config.Path, "/", 2)
	if len(pathParts) != 2 {
		return nil, fmt.Errorf("invalid ecr registry path")
	}
	pathParts2 := strings.Split(pathParts[0], ".")
	if len(pathParts2) != 6 {
		return nil, fmt.Errorf("invalid ecr registry path")
	}
	if pathParts2[1] != "dkr" || pathParts2[2] != "ecr" {
		return nil, fmt.Errorf("invalid ecr registry path")
	}

	return &ECRRegistryPath{
		AWSAccountID: pathParts2[0],
		Region:       pathParts2[3],
		Repo:         pathParts[1],
	}, nil
}

// GetTags filters and gets newer tags than the specified current tag
func (e *ECRRegistry) GetTags(currentTag string) ([]version.ImageVersion, error) {

	path, err := e.parseRegistryPath()
	if err != nil {
		return nil, err
	}

	c := e.Config
	log := c.Log.WithValues("image_repo", path.Repo)
	log.V(1).Info("GetTags", "account_id", path.AWSAccountID, "region", path.Region, "repo", path.Repo)

	// create aws session
	var sess *session.Session
	switch {
	case c.AWSCred.Profile != "":
		sess, err = session.NewSessionWithOptions(session.Options{
			Profile:           c.AWSCred.Profile,
			SharedConfigState: session.SharedConfigEnable,
		})
		log.V(1).Info("session created with profile", "profile_name", c.AWSCred.Profile)
	default:
		sess, err = session.NewSession(&aws.Config{Region: aws.String(path.Region)})
		log.V(1).Info("session created", "region", path.Region)
	}
	if err != nil {
		return nil, err
	}
	ecrSvc := ecr.New(sess)

	var nextToken *string
	imageVers := make([]version.ImageVersion, 0)
	for {
		// get all tags
		listImagesInput := ecr.ListImagesInput{
			RepositoryName: aws.String(path.Repo),
			RegistryId:     aws.String(path.AWSAccountID),
			MaxResults:     aws.Int64(100),
			NextToken:      nextToken,
		}
		listImagesOutput, err := ecrSvc.ListImages(&listImagesInput)
		if err != nil {
			return nil, err
		}

		// filter tags
		for _, imageId := range listImagesOutput.ImageIds {
			imageVer, err := version.NewImageVersion(*imageId.ImageTag, c.TagFormat)
			if err != nil {
				continue
			}
			if currentTag != "" {
				result, err := imageVer.Compare(currentTag)
				if err != nil {
					// occurs when the current version format changes
					log.Info(err.Error())
				}
				if result <= 0 {
					continue
				}
			}
			log.V(1).Info("new version found", "tag", imageVer.GetTag())
			imageVers = append(imageVers, imageVer)
		}

		if listImagesOutput.NextToken == nil {
			break
		}
		nextToken = listImagesOutput.NextToken
	}
	return imageVers, nil
}
