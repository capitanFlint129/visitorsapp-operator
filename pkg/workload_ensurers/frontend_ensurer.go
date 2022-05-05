package workload_ensurers

import (
	"context"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appv1alpha1 "example.com/m/v2/pkg/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

type frontendEnsurer struct {
	client client.Client
}

func (f *frontendEnsurer) EnsureDeployment(
	request reconcile.Request,
	instance *appv1alpha1.VisitorsApp,
	scheme *runtime.Scheme,
) (*reconcile.Result, error) {
	return ensureDeployment(request, instance, f.frontendDeployment(instance, scheme), f.client)
}

func (f *frontendEnsurer) EnsureService(
	request reconcile.Request,
	instance *appv1alpha1.VisitorsApp,
	scheme *runtime.Scheme,
) (*reconcile.Result, error) {
	return ensureService(request, instance, f.frontendService(instance, scheme), f.client)
}

func (f *frontendEnsurer) EnsureSecret(
	request reconcile.Request,
	instance *appv1alpha1.VisitorsApp,
	scheme *runtime.Scheme,
) (*reconcile.Result, error) {
	// TODO нужна ли тут паника или что-то еще?
	return nil, nil
}

const frontendPort = 3000
const frontendServicePort = 30686
const frontendImage = "jdob/visitors-webui:1.0.0"

func frontendDeploymentName(v *appv1alpha1.VisitorsApp) string {
	return v.Name + "-frontend"
}

func frontendServiceName(v *appv1alpha1.VisitorsApp) string {
	return v.Name + "-frontend-service"
}

func (f *frontendEnsurer) frontendDeployment(v *appv1alpha1.VisitorsApp, scheme *runtime.Scheme) *appsv1.Deployment {
	labels := labels(v, "frontend")
	size := int32(1)

	// If the header was specified, add it as an env variable
	env := []corev1.EnvVar{}
	if v.Spec.Title != "" {
		env = append(env, corev1.EnvVar{
			Name:  "REACT_APP_TITLE",
			Value: v.Spec.Title,
		})
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      frontendDeploymentName(v),
			Namespace: v.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &size,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: frontendImage,
						Name:  "visitors-webui",
						Ports: []corev1.ContainerPort{{
							ContainerPort: frontendPort,
							Name:          "visitors",
						}},
						Env: env,
					}},
				},
			},
		},
	}

	controllerutil.SetControllerReference(v, dep, scheme)
	return dep
}

func (f *frontendEnsurer) frontendService(v *appv1alpha1.VisitorsApp, scheme *runtime.Scheme) *corev1.Service {
	labels := labels(v, "frontend")

	s := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      frontendServiceName(v),
			Namespace: v.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{{
				Protocol:   corev1.ProtocolTCP,
				Port:       frontendPort,
				TargetPort: intstr.FromInt(frontendPort),
				NodePort:   frontendServicePort,
			}},
			Type: corev1.ServiceTypeNodePort,
		},
	}

	//log.Info("Service Spec", "Service.Name", s.ObjectMeta.Name)

	controllerutil.SetControllerReference(v, s, scheme)
	return s
}

func (f *frontendEnsurer) CheckWorkload(v *appv1alpha1.VisitorsApp) bool {
	// TODO подумать над реализацией
	return true
}

func (f *frontendEnsurer) UpdateStatus(instance *appv1alpha1.VisitorsApp) error {
	instance.Status.FrontendImage = frontendImage
	err := f.client.Status().Update(context.TODO(), instance)
	return err
}

func (f *frontendEnsurer) HandleWorkloadChanges(
	instance *appv1alpha1.VisitorsApp,
) (*reconcile.Result, error) {
	found := &appsv1.Deployment{}
	err := f.client.Get(context.TODO(), types.NamespacedName{
		Name:      frontendDeploymentName(instance),
		Namespace: instance.Namespace,
	}, found)
	if err != nil {
		// The deployment may not have been created yet, so requeue
		return &reconcile.Result{RequeueAfter: 5 * time.Second}, err
	}

	title := instance.Spec.Title
	existing := (*found).Spec.Template.Spec.Containers[0].Env[0].Value

	if title != existing {
		(*found).Spec.Template.Spec.Containers[0].Env[0].Value = title
		err = f.client.Update(context.TODO(), found)
		if err != nil {
			//log.Error(err, "Failed to update Deployment.", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return &reconcile.Result{}, err
		}
		// Spec updated - return and requeue
		return &reconcile.Result{Requeue: true}, nil
	}

	return nil, nil
}

func NewFrontendEnsurer(cli client.Client) WorkloadEnsurer {
	return &frontendEnsurer{
		client: cli,
	}
}
