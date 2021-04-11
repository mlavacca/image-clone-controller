package main

import (
	"flag"
	"image-clone-controller/pkg/controllers"
	"image-clone-controller/pkg/imagesManagement"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = appsv1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var (
		backupRegistry string
		backupRepository string
	)

	flag.StringVar(&backupRegistry, "backup-registry", "", "The registry to use as backup")
	flag.StringVar(&backupRepository, "backup-repository", "", "The repository to use as backup")
	klog.InitFlags(nil)
	flag.Parse()

	if backupRegistry == "" {
		klog.Error("--backup-registry flag is mandatory")
		os.Exit(1)
	}
	if backupRepository == "" {
		klog.Error("--backup-repository flag is mandatory")
		os.Exit(1)
	}

	if err := imagesManagement.SetupRegistryManager(backupRegistry, backupRepository); err != nil {
		klog.Fatal(err)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
	})
	if err != nil {
		klog.Fatal(err, " - unable to start manager")
	}

	if err = (&controllers.DeploymentReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		klog.Fatal(err, " - unable to create deployment controller")
	}

	if err = (&controllers.DaemonsetReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		klog.Fatal(err, " - unable to create daemonset controller")
	}

	klog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		klog.Fatal(err, " - problem while running manager")
	}
}
