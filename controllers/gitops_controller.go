package controllers

import (
	"context"
	"sort"

	"github.com/kazylla/gitops-controller/controllers/git"

	"github.com/kazylla/gitops-controller/controllers/registry"
	"github.com/kazylla/gitops-controller/controllers/version"

	"k8s.io/client-go/tools/record"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gitopsv1 "github.com/kazylla/gitops-controller/api/v1"
)

// GitOpsReconciler reconciles a GitOps object
type GitOpsReconciler struct {
	client.Client
	Log         logr.Logger
	Scheme      *runtime.Scheme
	GitUsername string
	GitPassword string
	Recorder    record.EventRecorder
}

// +kubebuilder:rbac:groups=gitops.kazylla.jp,resources=gitops,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=gitops.kazylla.jp,resources=gitops/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile periodically gets a new tag from the ECR and updates the image tag in the git repository with that tag
func (r *GitOpsReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("gitops", req.NamespacedName)

	var gitOps gitopsv1.GitOps
	log.Info("fetching GitOps Resource")
	if err := r.Get(ctx, req.NamespacedName, &gitOps); err != nil {
		log.Info("unable to fetch GitOps (maybe deleted)")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.V(1).Info("GitOps Resource",
		"image_path", gitOps.Spec.ImagePath,
		"image_tag_format", gitOps.Spec.ImageTagFormat,
		"git_repo", gitOps.Spec.GitRepo,
		"git_branch", gitOps.Spec.GitBranch,
		"git_release_branch", gitOps.Spec.GitReleaseBranch,
		"git_paths", gitOps.Spec.GitPaths,
		"git_commit_name", gitOps.Spec.GitCommitName,
		"git_commit_email", gitOps.Spec.GitCommitEmail,
	)

	// convert tag format
	var tagFmt version.TagFormat
	switch gitOps.Spec.ImageTagFormat {
	case "serial":
		tagFmt = version.TagFormatSerial
	case "semantic":
		tagFmt = version.TagFormatSemantic
	default:
		log.Info("invalid tag format", "format", gitOps.Spec.ImageTagFormat)
		return ctrl.Result{}, nil
	}

	// get filtered tags
	log.V(1).Info("scanning docker registry", "current_tag", gitOps.Status.CurrentTag)
	ecrRegistry := registry.NewRegistry(registry.Config{
		Path:      gitOps.Spec.ImagePath,
		TagFormat: tagFmt,
		Log:       log,
		AWSCred: registry.AWSCred{
			Profile: gitOps.Spec.AWSProfile,
		},
	})

	imageVers, err := ecrRegistry.GetTags(gitOps.Status.CurrentTag)
	if err != nil {
		return ctrl.Result{}, err
	}

	// sort image version by ascending
	sort.Slice(imageVers, func(i, j int) bool {
		result, _ := imageVers[i].Compare(imageVers[j].GetTag())
		return result < 0
	})
	log.V(1).Info("scanning docker registry has succeeded", "new", len(imageVers))

	if len(imageVers) == 0 {
		return ctrl.Result{}, nil
	}

	// commit uncommitted tags from oldest
	gitRepo, err := git.NewGitRepo(git.Config{
		ImagePath:     gitOps.Spec.ImagePath,
		Repo:          gitOps.Spec.GitRepo,
		Branch:        gitOps.Spec.GitBranch,
		ReleaseBranch: gitOps.Spec.GitReleaseBranch,
		Paths:         gitOps.Spec.GitPaths,
		CommitName:    gitOps.Spec.GitCommitName,
		CommitEmail:   gitOps.Spec.GitCommitEmail,
		Username:      r.GitUsername,
		Password:      r.GitPassword,
		Log:           log,
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	latestTag, err := gitRepo.CommitTags(imageVers)
	if err != nil {
		return ctrl.Result{}, err
	}

	// update CurrentTag status to latest tag
	if gitOps.Status.CurrentTag != latestTag {

		log.Info("all uncommited tags has commited", "latest_tag", latestTag)
		gitOps.Status.CurrentTag = latestTag

		// update gitops.status
		if err := r.Status().Update(ctx, &gitOps); err != nil {
			log.Error(err, "unable to update GitOps status")
			return ctrl.Result{}, err
		}

		// create event for updated gitops.status
		r.Recorder.Eventf(&gitOps, corev1.EventTypeNormal, "Updated", "Update gitops.status.current_tag: %s", gitOps.Status.CurrentTag)
	}

	return ctrl.Result{}, nil
}

func (r *GitOpsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gitopsv1.GitOps{}).
		Complete(r)
}
