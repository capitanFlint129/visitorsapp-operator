package workload_ensurers

import (
	"context"
	"time"

	appv1alpha1 "example.com/m/v2/pkg/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type backendEnsurer struct {
	client            client.Client
	port              int
	servicePort       int
	image             string
	mysqlAuthName     string
	mysqlServiceName  string
	deploymentPostfix string
	servicePostfix    string
}

func (b *backendEnsurer) EnsureDeployment(
	request reconcile.Request,
	instance *appv1alpha1.VisitorsApp,
	scheme *runtime.Scheme,
) (*reconcile.Result, error) {
	return ensureDeployment(request, instance, b.backendDeployment(instance, scheme), b.client)
}

func (b *backendEnsurer) EnsureService(
	request reconcile.Request,
	instance *appv1alpha1.VisitorsApp,
	scheme *runtime.Scheme,
) (*reconcile.Result, error) {
	return ensureService(request, instance, b.backendService(instance, scheme), b.client)
}

func (b *backendEnsurer) EnsureSecret(
	request reconcile.Request,
	instance *appv1alpha1.VisitorsApp,
	scheme *runtime.Scheme,
) (*reconcile.Result, error) {
	return nil, nil
}

func (b *backendEnsurer) CheckWorkload(v *appv1alpha1.VisitorsApp) bool {
	return true
}

func (b *backendEnsurer) UpdateStatus(instance *appv1alpha1.VisitorsApp) error {
	instance.Status.BackendImage = b.image
	err := b.client.Status().Update(context.TODO(), instance)
	return err
}

func (b *backendEnsurer) HandleWorkloadChanges(
	instance *appv1alpha1.VisitorsApp,
) (*reconcile.Result, error) {
	found := &appsv1.Deployment{}
	err := b.client.Get(context.TODO(), types.NamespacedName{
		Name:      instance.Name + b.deploymentPostfix,
		Namespace: instance.Namespace,
	}, found)
	if err != nil {
		// The deployment may not have been created yet, so requeue
		return &reconcile.Result{RequeueAfter: 5 * time.Second}, err
	}

	size := instance.Spec.Size

	if size != *found.Spec.Replicas {
		found.Spec.Replicas = &size
		err = b.client.Update(context.TODO(), found)
		if err != nil {
			//log.Error(err, "Failed to update Deployment.", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return &reconcile.Result{}, err
		}
		// Spec updated - return and requeue
		return &reconcile.Result{Requeue: true}, nil
	}

	return nil, nil
}

func (b *backendEnsurer) backendDeployment(v *appv1alpha1.VisitorsApp, scheme *runtime.Scheme) *appsv1.Deployment {
	labels := labels(v, "backend")
	size := v.Spec.Size

	userSecret := &corev1.EnvVarSource{
		SecretKeyRef: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: b.mysqlAuthName},
			Key:                  "username",
		},
	}

	passwordSecret := &corev1.EnvVarSource{
		SecretKeyRef: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: b.mysqlAuthName},
			Key:                  "password",
		},
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      v.Name + b.deploymentPostfix,
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
						Image:           b.image,
						ImagePullPolicy: corev1.PullNever,
						Name:            "visitors-service",
						Ports: []corev1.ContainerPort{{
							ContainerPort: int32(b.port),
							Name:          "visitors",
						}},
						Env: []corev1.EnvVar{
							{
								Name:  "MYSQL_DATABASE",
								Value: "visitors",
							},
							{
								Name:  "MYSQL_SERVICE_HOST",
								Value: b.mysqlServiceName,
							},
							{
								Name:      "MYSQL_USERNAME",
								ValueFrom: userSecret,
							},
							{
								Name:      "MYSQL_PASSWORD",
								ValueFrom: passwordSecret,
							},
						},
					}},
				},
			},
		},
	}

	controllerutil.SetControllerReference(v, dep, scheme)
	return dep
}

func (b *backendEnsurer) backendService(v *appv1alpha1.VisitorsApp, scheme *runtime.Scheme) *corev1.Service {
	labels := labels(v, "backend")

	s := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      v.Name + b.servicePostfix,
			Namespace: v.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{{
				Protocol:   corev1.ProtocolTCP,
				Port:       int32(b.port),
				TargetPort: intstr.FromInt(b.port),
				NodePort:   30685,
			}},
			Type: corev1.ServiceTypeNodePort,
		},
	}

	controllerutil.SetControllerReference(v, s, scheme)
	return s
}

func NewBackendEnsurer(
	cli client.Client,
	port int,
	servicePort int,
	image string,
	mysqlAuthName string,
	mysqlServiceName string,
	deploymentPostfix string,
	servicePostfix string,
) WorkloadEnsurer {
	return &backendEnsurer{
		client:            cli,
		port:              port,
		servicePort:       servicePort,
		image:             image,
		mysqlAuthName:     mysqlAuthName,
		mysqlServiceName:  mysqlServiceName,
		deploymentPostfix: deploymentPostfix,
		servicePostfix:    servicePostfix,
	}
}
