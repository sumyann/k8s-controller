/*
Copyright 2023.

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

package controller

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appv1alpha1 "github.com/sumyann/k8s-controller/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// MyAppResourceReconciler reconciles a MyAppResource object
type MyAppResourceReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=app.example.com,resources=myappresources,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=app.example.com,resources=myappresources/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

func (r *MyAppResourceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("myappresource", req.NamespacedName)
	log.Info("Starting reconciliation", "namespace", req.NamespacedName.Namespace, "name", req.NamespacedName.Name)

	// Fetch the MyAppResource instance
	myAppResource := &appv1alpha1.MyAppResource{}
	err := r.Get(ctx, req.NamespacedName, myAppResource)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return. Created objects are automatically garbage collected.
			log.Info("MyAppResource resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get MyAppResource")
		return ctrl.Result{}, err
	}

	// Define a new Podinfo deployment
	podinfoDeployment := r.deploymentForPodinfo(myAppResource)
	// Set MyAppResource instance as the owner and controller
	ctrl.SetControllerReference(myAppResource, podinfoDeployment, r.Scheme)

	// Check if this Podinfo Deployment already exists
	found := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: podinfoDeployment.Name, Namespace: podinfoDeployment.Namespace}, found)

	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new Deployment", "Deployment.Namespace", podinfoDeployment.Namespace, "Deployment.Name", podinfoDeployment.Name)
		err = r.Create(ctx, podinfoDeployment)
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", podinfoDeployment.Namespace, "Deployment.Name", podinfoDeployment.Name)
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	} else {
		// Update the deployment if necessary
		updateNeeded := false

		// Check and update the replica count
		replicaCount := myAppResource.Spec.ReplicaCount
		if *found.Spec.Replicas != replicaCount {
			found.Spec.Replicas = &replicaCount
			updateNeeded = true
		}

		// Check and update the environment variables
		mergedEnvVars := r.mergeEnvVars(myAppResource, found)
		if !equalEnvVars(found.Spec.Template.Spec.Containers[0].Env, mergedEnvVars) {
			found.Spec.Template.Spec.Containers[0].Env = mergedEnvVars
			updateNeeded = true
		}

		// If an update is needed, update the deployment
		if updateNeeded {
			log.Info("Updating Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			err = r.Update(ctx, found)
			if err != nil {
				log.Error(err, "Failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
				return ctrl.Result{}, err
			}
			return ctrl.Result{Requeue: true}, nil
		}
	}

	// Update the MyAppResource status with the pod names
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(myAppResource.Namespace),
		client.MatchingLabels(labelsForPodinfo(myAppResource.Name)),
	}
	if err = r.List(ctx, podList, listOpts...); err != nil {
		log.Error(err, "Failed to list pods", "MyAppResource.Namespace", myAppResource.Namespace, "MyAppResource.Name", myAppResource.Name)
		return ctrl.Result{}, err
	}

	log.Info("Ending reconciliation", "namespace", req.NamespacedName.Namespace, "name", req.NamespacedName.Name)

	return ctrl.Result{}, nil
}

// mergeEnvVars merges the environment variables from the CR and the existing deployment,
// updating or appending the environment variables from the CR.
func (r *MyAppResourceReconciler) mergeEnvVars(m *appv1alpha1.MyAppResource, d *appsv1.Deployment) []corev1.EnvVar {
	// Create a map to hold the final set of environment variables
	envVarMap := make(map[string]corev1.EnvVar)

	// Populate the map with environment variables from the existing deployment
	for _, envVar := range d.Spec.Template.Spec.Containers[0].Env {
		envVarMap[envVar.Name] = envVar
	}

	// Update the map with environment variables from the CR, which will
	// overwrite any existing environment variables with the same name
	for _, envVar := range m.Spec.Env {
		envVarMap[envVar.Name] = envVar
	}

	// Convert the map back to a slice
	mergedEnvVars := make([]corev1.EnvVar, 0, len(envVarMap))
	for _, envVar := range envVarMap {
		mergedEnvVars = append(mergedEnvVars, envVar)
	}

	return mergedEnvVars
}

// equalEnvVars checks if two slices of environment variables are equal.
func equalEnvVars(envVars1, envVars2 []corev1.EnvVar) bool {
	if len(envVars1) != len(envVars2) {
		return false
	}

	// Convert slices to maps for easier comparison
	envVarMap1 := make(map[string]string, len(envVars1))
	envVarMap2 := make(map[string]string, len(envVars2))

	for _, envVar := range envVars1 {
		envVarMap1[envVar.Name] = envVar.Value
	}
	for _, envVar := range envVars2 {
		envVarMap2[envVar.Name] = envVar.Value
	}

	// Compare maps
	for key, value := range envVarMap1 {
		if envVarMap2[key] != value {
			return false
		}
	}

	return true
}

func (r *MyAppResourceReconciler) deploymentForPodinfo(m *appv1alpha1.MyAppResource) *appsv1.Deployment {
	labels := labelsForPodinfo(m.Name)

	// Merge environment variables from the env field with other environment variables
	envVars := append(m.Spec.Env, []corev1.EnvVar{
		{
			Name:  "PODINFO_UI_COLOR",
			Value: m.Spec.UI.Color,
		},
		{
			Name:  "PODINFO_UI_MESSAGE",
			Value: m.Spec.UI.Message,
		},
	}...)

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name + "-podinfo",
			Namespace: m.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &m.Spec.ReplicaCount,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "podinfo",
							Image: "ghcr.io/stefanprodan/podinfo:latest",
							Env:   envVars, // to use merged environment variables
						},
					},
				},
			},
		},
	}
}

func labelsForPodinfo(name string) map[string]string {
	return map[string]string{"app": "podinfo", "podinfo_cr": name}
}

func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}

func (r *MyAppResourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1alpha1.MyAppResource{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}

func (r *MyAppResourceReconciler) deploymentForRedis(m *appv1alpha1.MyAppResource) *appsv1.Deployment {
	labels := labelsForRedis(m.Name)
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name + "-redis",
			Namespace: m.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &m.Spec.ReplicaCount, // assuming Redis should have same replica count, adjust as needed
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "redis",
							Image: "redis:latest",
						},
					},
				},
			},
		},
	}
}

func labelsForRedis(name string) map[string]string {
	return map[string]string{"app": "redis", "redis_cr": name}
}
