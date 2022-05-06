package workload_ensurers

import (
	"time"

	appv1alpha1 "example.com/m/v2/pkg/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ensureWorkloadDirector struct {
	ensurer WorkloadEnsurer
}

func (e *ensureWorkloadDirector) SetEnsurer(ensurer WorkloadEnsurer) {
	e.ensurer = ensurer
}

func (e *ensureWorkloadDirector) EnsureMysql(
	request reconcile.Request,
	instance *appv1alpha1.VisitorsApp,
	scheme *runtime.Scheme,
) (*reconcile.Result, error) {
	result, err := e.ensurer.EnsureSecret(request, instance, scheme)
	if result != nil {
		return result, err
	}

	result, err = e.ensurer.EnsureDeployment(request, instance, scheme)
	if result != nil {
		return result, err
	}

	result, err = e.ensurer.EnsureService(request, instance, scheme)
	if result != nil {
		return result, err
	}

	mysqlRunning := e.ensurer.CheckWorkload(instance)

	if !mysqlRunning {
		// If MySQL isn't running yet, requeue the reconcile
		// to run again after a delay
		delay := time.Second * time.Duration(5)

		// log.Info(fmt.Sprintf("MySQL isn't running, waiting for %s", delay))
		return &reconcile.Result{RequeueAfter: delay}, nil
	}
	return nil, nil
}

func (e *ensureWorkloadDirector) EnsureBackend(
	request reconcile.Request,
	instance *appv1alpha1.VisitorsApp,
	scheme *runtime.Scheme,
) (*reconcile.Result, error) {
	result, err := e.ensurer.EnsureDeployment(request, instance, scheme)
	if result != nil {
		return result, err
	}

	result, err = e.ensurer.EnsureService(request, instance, scheme)
	if result != nil {
		return result, err
	}

	err = e.ensurer.UpdateStatus(instance)
	if err != nil {
		// Requeue the request if the status could not be updated
		return &reconcile.Result{}, err
	}

	result, err = e.ensurer.HandleWorkloadChanges(instance)
	if result != nil {
		return result, err
	}
	return nil, nil
}

func (e *ensureWorkloadDirector) EnsureFrontend(
	request reconcile.Request,
	instance *appv1alpha1.VisitorsApp,
	scheme *runtime.Scheme,
) (*reconcile.Result, error) {
	result, err := e.ensurer.EnsureDeployment(request, instance, scheme)
	if result != nil {
		return result, err
	}

	result, err = e.ensurer.EnsureService(request, instance, scheme)
	if result != nil {
		return result, err
	}

	err = e.ensurer.UpdateStatus(instance)
	if err != nil {
		// Requeue the request
		return &reconcile.Result{}, err
	}

	result, err = e.ensurer.HandleWorkloadChanges(instance)
	if result != nil {
		return result, err
	}
	return nil, nil
}

func NewEnsureWorkloadDirector() EnsureWorkloadDirector {
	return &ensureWorkloadDirector{}
}
