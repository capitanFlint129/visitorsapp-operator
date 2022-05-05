package controllers

import (
	"context"
	appv1alpha1 "example.com/m/v2/pkg/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ensureWorkloadDirector = interface {
	SetEnsurer(ensurer workloadEnsurer)
	EnsureMysql(
		request reconcile.Request,
		instance *appv1alpha1.VisitorsApp,
		scheme *runtime.Scheme,
	) (*reconcile.Result, error)
	EnsureBackend(
		request reconcile.Request,
		instance *appv1alpha1.VisitorsApp,
		scheme *runtime.Scheme,
	) (*reconcile.Result, error)
	EnsureFrontend(
		request reconcile.Request,
		instance *appv1alpha1.VisitorsApp,
		scheme *runtime.Scheme,
	) (*reconcile.Result, error)
}

type workloadEnsurer = interface {
	EnsureDeployment(
		request reconcile.Request,
		instance *appv1alpha1.VisitorsApp,
		scheme *runtime.Scheme,
	) (*reconcile.Result, error)
	EnsureService(
		request reconcile.Request,
		instance *appv1alpha1.VisitorsApp,
		scheme *runtime.Scheme,
	) (*reconcile.Result, error)
	EnsureSecret(
		request reconcile.Request,
		instance *appv1alpha1.VisitorsApp,
		scheme *runtime.Scheme,
	) (*reconcile.Result, error)
	UpdateStatus(instance *appv1alpha1.VisitorsApp) error
	HandleWorkloadChanges(
		instance *appv1alpha1.VisitorsApp,
	) (*reconcile.Result, error)
	CheckWorkload(instance *appv1alpha1.VisitorsApp) bool
}

type Controller interface {
	Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error)
	SetupWithManager(mgr ctrl.Manager) error
}
