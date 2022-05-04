package controllers

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
)

type ensureWorkloadDirector = interface {
	SetEnsurer(ensurer workloadEnsurer)
	EnsureMySQL()
	EnsureBackend()
	EnsureFrontend()
}

type workloadEnsurer = interface {
	EnsureDeployment()
	EnsureService()
	EnsureSecret()
}

type Controller interface {
	Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error)
	SetupWithManager(mgr ctrl.Manager) error
}
