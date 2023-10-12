package controller

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	appv1alpha1 "github.com/sumyann/k8s-controller/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

func TestMyAppResourceReconciler_Reconcile(t *testing.T) {
	// Create a new MyAppResource custom resource
	exampleResource := &appv1alpha1.MyAppResource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-app",
			Namespace: "default",
		},
		Spec: appv1alpha1.MyAppResourceSpec{},
	}

	// Create the resource in the Kubernetes cluster
	err := k8sClient.Create(context.TODO(), exampleResource)
	require.NoError(t, err)

	// Create a new reconciler
	r := &MyAppResourceReconciler{
		Client: k8sClient,
		Log:    ctrl.Log.WithName("controllers").WithName("MyAppResource"),
		Scheme: scheme,
	}

	// Call the Reconcile method
	_, err = r.Reconcile(context.TODO(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "example-app",
			Namespace: "default",
		},
	})
	require.NoError(t, err)

}
