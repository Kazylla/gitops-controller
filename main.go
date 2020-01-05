package main

import (
	"flag"
	"os"
	"regexp"
	"strconv"
	"time"

	gitopsv1 "github.com/kazylla/gitops-controller/api/v1"
	"github.com/kazylla/gitops-controller/controllers"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = gitopsv1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	// set variables from env
	rep := regexp.MustCompile(".")
	envDevelopment := getenv("GITOPS_DEVELOPMENT", "false")
	envResyncPeriod := getenvInt("GITOPS_RESYNC_PERIOD", 30)
	envGitUsername := getenv("GITOPS_GIT_USERNAME", "")
	envGitPassword := getenv("GITOPS_GIT_PASSWORD", "")

	ctrl.SetLogger(zap.New(func(o *zap.Options) {
		o.Development = envDevelopment == "true"
	}))

	// log variables from env
	setupLog.Info("GITOPS_DEVELOPMENT", "value", envDevelopment)
	setupLog.Info("GITOPS_RESYNC_PERIOD", "value", envResyncPeriod)
	setupLog.Info("GITOPS_GIT_USERNAME", "value", envGitUsername)
	setupLog.Info("GITOPS_GIT_PASSWORD", "value", rep.ReplaceAllString(envGitPassword, "*"))

	var resyncPeriod = time.Second * time.Duration(envResyncPeriod)

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		SyncPeriod:         &resyncPeriod,
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		LeaderElection:     enableLeaderElection,
		Port:               9443,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.GitOpsReconciler{
		Client:      mgr.GetClient(),
		Log:         ctrl.Log.WithName("controllers").WithName("GitOps"),
		Scheme:      mgr.GetScheme(),
		GitUsername: envGitUsername,
		GitPassword: envGitPassword,
		Recorder:    mgr.GetEventRecorderFor("gitops-controller"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GitOps")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func getenvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return intValue
}
