package workload_ensurers

import (
	"context"

	appv1alpha1 "example.com/m/v2/pkg/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func ensureDeployment(
	request reconcile.Request,
	instance *appv1alpha1.VisitorsApp,
	dep *appsv1.Deployment,
	cli client.Client,
) (*reconcile.Result, error) {

	// See if deployment already exists and create if it doesn't
	found := &appsv1.Deployment{}
	err := cli.Get(context.TODO(), types.NamespacedName{
		Name:      dep.Name,
		Namespace: instance.Namespace,
	}, found)
	if err != nil && errors.IsNotFound(err) {

		// Create the deployment
		//log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = cli.Create(context.TODO(), dep)

		if err != nil {
			// Deployment failed
			//log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return &reconcile.Result{}, err
		} else {
			// Deployment was successful
			return nil, nil
		}
	} else if err != nil {
		// Error that isn't due to the deployment not existing
		//log.Error(err, "Failed to get Deployment")
		return &reconcile.Result{}, err
	}

	return nil, nil
}

func ensureService(
	request reconcile.Request,
	instance *appv1alpha1.VisitorsApp,
	s *corev1.Service,
	cli client.Client,
) (*reconcile.Result, error) {
	found := &corev1.Service{}
	err := cli.Get(context.TODO(), types.NamespacedName{
		Name:      s.Name,
		Namespace: instance.Namespace,
	}, found)
	if err != nil && errors.IsNotFound(err) {

		// Create the service
		//log.Info("Creating a new Service", "Service.Namespace", s.Namespace, "Service.Name", s.Name)
		err = cli.Create(context.TODO(), s)

		if err != nil {
			// Creation failed
			//log.Error(err, "Failed to create new Service", "Service.Namespace", s.Namespace, "Service.Name", s.Name)
			return &reconcile.Result{}, err
		} else {
			// Creation was successful
			return nil, nil
		}
	} else if err != nil {
		// Error that isn't due to the service not existing
		//log.Error(err, "Failed to get Service")
		return &reconcile.Result{}, err
	}

	return nil, nil
}

func ensureSecret(request reconcile.Request,
	instance *appv1alpha1.VisitorsApp,
	s *corev1.Secret,
	cli client.Client,
) (*reconcile.Result, error) {
	found := &corev1.Secret{}
	err := cli.Get(context.TODO(), types.NamespacedName{
		Name:      s.Name,
		Namespace: instance.Namespace,
	}, found)
	if err != nil && errors.IsNotFound(err) {
		// Create the secret
		//log.Info("Creating a new secret", "Secret.Namespace", s.Namespace, "Secret.Name", s.Name)
		err = cli.Create(context.TODO(), s)

		if err != nil {
			// Creation failed
			//log.Error(err, "Failed to create new Secret", "Secret.Namespace", s.Namespace, "Secret.Name", s.Name)
			return &reconcile.Result{}, err
		} else {
			// Creation was successful
			return nil, nil
		}
	} else if err != nil {
		// Error that isn't due to the secret not existing
		//log.Error(err, "Failed to get Secret")
		return &reconcile.Result{}, err
	}

	return nil, nil
}

func labels(v *appv1alpha1.VisitorsApp, tier string) map[string]string {
	return map[string]string{
		"app":             "visitors",
		"visitorssite_cr": v.Name,
		"tier":            tier,
	}
}
