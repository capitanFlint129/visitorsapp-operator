/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// TODO добавить pkg
package controllers

import (
	"context"
	appv1alpha1 "example.com/m/v2/pkg/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var log = logf.Log.WithName("controller_visitorsapp")

// VisitorsAppController reconciles a VisitorsApp object
type VisitorsAppController struct {
	Client                 client.Client
	Scheme                 *runtime.Scheme
	ensureWorkloadDirector ensureWorkloadDirector
	mysqlEnsurer           workloadEnsurer
	backendEnsurer         workloadEnsurer
	frontendEnsurer        workloadEnsurer
}

//+kubebuilder:rbac:groups=app.my.domain,resources=visitorsapps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=app.my.domain,resources=visitorsapps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=app.my.domain,resources=visitorsapps/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the VisitorsApp object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *VisitorsAppController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	//reqLogger := log.WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	//reqLogger.Info("Reconciling VisitorsApp")

	// Fetch the VisitorsApp instance
	visitorAppInstance := &appv1alpha1.VisitorsApp{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, visitorAppInstance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	var result *reconcile.Result

	// == MySQL ==========
	r.ensureWorkloadDirector.SetEnsurer(r.mysqlEnsurer)
	result, err = r.ensureWorkloadDirector.EnsureMysql(req, visitorAppInstance, r.Scheme)
	if result != nil {
		return *result, err
	}

	// == Visitors Backend  ==========
	r.ensureWorkloadDirector.SetEnsurer(r.backendEnsurer)
	result, err = r.ensureWorkloadDirector.EnsureBackend(req, visitorAppInstance, r.Scheme)
	if result != nil {
		return *result, err
	}

	// == Visitors Frontend ==========
	r.ensureWorkloadDirector.SetEnsurer(r.frontendEnsurer)
	result, err = r.ensureWorkloadDirector.EnsureFrontend(req, visitorAppInstance, r.Scheme)
	if result != nil {
		return *result, err
	}

	// == Finish ==========
	// Everything went fine, don't requeue
	return reconcile.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *VisitorsAppController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1alpha1.VisitorsApp{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}

func NewVisitorsAppController(
	cli client.Client,
	scheme *runtime.Scheme,
	ensureWorkloadDirector ensureWorkloadDirector,
	mysqlEnsurer workloadEnsurer,
	backendEnsurer workloadEnsurer,
	frontendEnsurer workloadEnsurer,
) Controller {
	return &VisitorsAppController{
		Client:                 cli,
		Scheme:                 scheme,
		ensureWorkloadDirector: ensureWorkloadDirector,
		mysqlEnsurer:           mysqlEnsurer,
		backendEnsurer:         backendEnsurer,
		frontendEnsurer:        frontendEnsurer,
	}
}
