/*
Copyright 2024.

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

package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appv1 "mfabriczy/clientapp-operator/api/v1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// ClientAppReconciler reconciles a ClientApp object
type ClientAppReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=app.mfabriczy,resources=clientapps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=app.mfabriczy,resources=clientapps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=app.mfabriczy,resources=clientapps/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ClientApp object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *ClientAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	clientApp := &appv1.ClientApp{}

	// Fetch CR instance.
	err := r.Get(ctx, req.NamespacedName, clientApp)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("clientApp resource not found", "namespace", req.NamespacedName.Namespace, "name", req.NamespacedName.Name)
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get clientApp resource", "namespace", req.NamespacedName.Namespace, "name", req.NamespacedName.Name)
		return ctrl.Result{}, err
	}

	// Get deployment from the same namespace as the clientApp CR.
	deployment := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: clientApp.Name, Namespace: clientApp.Namespace}, deployment)
	if err != nil {
		if errors.IsNotFound(err) {
			// Define and create a new deployment.
			log.Info("clientApp deployment not found, creating", "name", clientApp.Name, "namespace", clientApp.Namespace)
			dep := r.clientAppDeployment(clientApp)
			if err = r.Create(ctx, dep); err != nil {
				log.Error(err, "Failed to create deployment for clientApp", "namespace", req.NamespacedName.Namespace, "name", req.NamespacedName.Name)
				return ctrl.Result{}, err
			}
			return ctrl.Result{Requeue: true}, nil
		} else {
			return ctrl.Result{}, err
		}
	}

	// Check if the replica count matches the CR.
	replicas := clientApp.Spec.Replicas
	if *deployment.Spec.Replicas != replicas {
		deployment.Spec.Replicas = &replicas
		err = r.Update(ctx, deployment)
		if err != nil {
			log.Error(err, "Failed to update Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
			return ctrl.Result{}, err
		}

		if err := r.Status().Update(ctx, clientApp); err != nil {
			log.Error(err, "Failed to update ClientApp status after replica change", "namespace", clientApp.Namespace, "name", clientApp.Name)
			return ctrl.Result{}, err
		}

		log.Info("Updated clientApp status with new replica count", "replicas", replicas)

		// Spec updated - return and requeue
		return ctrl.Result{Requeue: true}, nil
	}

	// Update CR status.
	availableCondition := metav1.Condition{
		Type:    "Available",
		Status:  metav1.ConditionFalse, // Assume not available by default
		Reason:  "DeploymentUnavailable",
		Message: "The deployment is not available",
	}

	for _, cond := range deployment.Status.Conditions {
		if cond.Type == appsv1.DeploymentAvailable && cond.Status == corev1.ConditionTrue {
			availableCondition.Status = metav1.ConditionTrue
			availableCondition.Reason = "DeploymentAvailable"
			availableCondition.Message = "The deployment is available and running"
			break // Condition found, exit loop.
		}
	}

	// Update the ClientApp status with the appropriate condition
	clientApp.Status.Conditions = []metav1.Condition{availableCondition}

	if err := r.Status().Update(ctx, clientApp); err != nil {
		log.Error(err, "Failed to update ClientApp status", "namespace", clientApp.Namespace, "name", clientApp.Name)
		return ctrl.Result{}, err // Returning an error will make sure the request is requeued for another try
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClientAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1.ClientApp{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}

func (r *ClientAppReconciler) clientAppDeployment(c *appv1.ClientApp) *appsv1.Deployment {
	labels := map[string]string{"cr_name": c.Spec.Name}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.Spec.Name,
			Namespace: c.Namespace, // Set to the same namespace as the CR.
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &c.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: c.Spec.Image,
						Name:  c.Spec.Name,
						Ports: []corev1.ContainerPort{{
							ContainerPort: c.Spec.Port,
							Name:          c.Spec.PortName,
						}},
						Env:       c.Spec.Env,
						Resources: c.Spec.Resources,
					}},
				},
			},
		},
	}

	// Set the clientApp CR as the owner of the Deployment for garbage collection purposes; for example, to clean up the Deployment when the CR is removed.
	controllerutil.SetControllerReference(c, dep, r.Scheme)
	return dep
}
