package workload_ensurers

import (
	"context"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appv1alpha1 "example.com/m/v2/pkg/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type mysqlEnsurer struct {
	client         client.Client
	deploymentName string
	serviceName    string
	authName       string
}

func (m *mysqlEnsurer) EnsureDeployment(
	request reconcile.Request,
	instance *appv1alpha1.VisitorsApp,
	scheme *runtime.Scheme,
) (*reconcile.Result, error) {
	return ensureDeployment(request, instance, m.mysqlDeployment(instance, scheme), m.client)
}

func (m *mysqlEnsurer) EnsureService(
	request reconcile.Request,
	instance *appv1alpha1.VisitorsApp,
	scheme *runtime.Scheme,
) (*reconcile.Result, error) {
	return ensureService(request, instance, m.mysqlService(instance, scheme), m.client)
}

func (m *mysqlEnsurer) EnsureSecret(
	request reconcile.Request,
	instance *appv1alpha1.VisitorsApp,
	scheme *runtime.Scheme,
) (*reconcile.Result, error) {
	return ensureSecret(request, instance, m.mysqlAuthSecret(instance, scheme), m.client)
}

// CheckWorkload returns whether the MySQL deployment is running
func (m *mysqlEnsurer) CheckWorkload(v *appv1alpha1.VisitorsApp) bool {
	deployment := &appsv1.Deployment{}

	err := m.client.Get(context.TODO(), types.NamespacedName{
		Name:      m.deploymentName,
		Namespace: v.Namespace,
	}, deployment)

	if err != nil {
		// log.Error(err, "Deployment mysql not found")
		return false
	}

	if deployment.Status.ReadyReplicas == 1 {
		return true
	}

	return false
}

func (m *mysqlEnsurer) UpdateStatus(instance *appv1alpha1.VisitorsApp) error {
	return nil
}

func (m *mysqlEnsurer) HandleWorkloadChanges(
	instance *appv1alpha1.VisitorsApp,
) (*reconcile.Result, error) {
	return nil, nil
}

func (m *mysqlEnsurer) mysqlAuthSecret(v *appv1alpha1.VisitorsApp, scheme *runtime.Scheme) *corev1.Secret {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.authName,
			Namespace: v.Namespace,
		},
		Type: "Opaque",
		StringData: map[string]string{
			"username": "visitors-user",
			"password": "visitors-pass",
		},
	}
	controllerutil.SetControllerReference(v, secret, scheme)
	return secret
}

func (m *mysqlEnsurer) mysqlDeployment(v *appv1alpha1.VisitorsApp, scheme *runtime.Scheme) *appsv1.Deployment {
	labels := labels(v, "mysql")
	size := v.Spec.Size

	userSecret := &corev1.EnvVarSource{
		SecretKeyRef: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: m.authName},
			Key:                  "username",
		},
	}

	passwordSecret := &corev1.EnvVarSource{
		SecretKeyRef: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: m.authName},
			Key:                  "password",
		},
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.deploymentName,
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
						Image: "mysql:5.7",
						Name:  "visitors-mysql",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 3306,
							Name:          "mysql",
						}},
						Env: []corev1.EnvVar{
							{
								Name:  "MYSQL_ROOT_PASSWORD",
								Value: "password",
							},
							{
								Name:  "MYSQL_DATABASE",
								Value: "visitors",
							},
							{
								Name:      "MYSQL_USER",
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

func (m *mysqlEnsurer) mysqlService(v *appv1alpha1.VisitorsApp, scheme *runtime.Scheme) *corev1.Service {
	labels := labels(v, "mysql")

	s := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.serviceName,
			Namespace: v.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{{
				Port: 3306,
			}},
			ClusterIP: "None",
		},
	}

	controllerutil.SetControllerReference(v, s, scheme)
	return s
}

func NewMysqlEnsurer(
	cli client.Client,
	deploymentName string,
	serviceName string,
	authName string,
) WorkloadEnsurer {
	return &mysqlEnsurer{
		client:         cli,
		deploymentName: deploymentName,
		serviceName:    serviceName,
		authName:       authName,
	}
}
