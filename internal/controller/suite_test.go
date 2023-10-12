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
	"log"
	"os"
	"testing"

	appv1alpha1 "github.com/sumyann/k8s-controller/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	scheme    = runtime.NewScheme()
	k8sClient client.Client
	k8sMgr    manager.Manager
	testEnv   *envtest.Environment
)

func boolPtr(b bool) *bool {
	return &b
}

func TestMain(m *testing.M) {
	clientgoscheme.AddToScheme(scheme)
	appv1alpha1.AddToScheme(scheme)

	testEnv = &envtest.Environment{
		UseExistingCluster: boolPtr(true),
	}

	var err error
	cfg, err := testEnv.Start()
	if err != nil {
		log.Fatal(err)
	}

	k8sMgr, err = manager.New(cfg, manager.Options{Scheme: scheme})
	if err != nil {
		log.Fatal(err)
	}

	k8sClient = k8sMgr.GetClient()

	ctrl.SetLogger(zap.New(zap.WriteTo(os.Stdout), zap.UseDevMode(true)))

	go func() {
		err = k8sMgr.Start(context.TODO())
		if err != nil {
			log.Fatal(err)
		}
	}()

	code := m.Run()

	err = testEnv.Stop()
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(code)
}
