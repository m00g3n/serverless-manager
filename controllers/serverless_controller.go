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

package controllers

import (
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	"github.com/kyma-project/module-manager/operator/pkg/declarative"
	"github.com/kyma-project/module-manager/operator/pkg/types"
	"github.com/kyma-project/serverless-manager/api/v1alpha1"
)

const (
	chartNs = "kyma-system"
)

// ServerlessReconciler reconciles a Serverless object
type ServerlessReconciler struct {
	declarative.ManifestReconciler
	client.Client
	Scheme *runtime.Scheme
	*rest.Config
	ChartPath string
}

//+kubebuilder:rbac:groups=operator.kyma-project.io,resources=serverlesses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operator.kyma-project.io,resources=serverlesses/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=operator.kyma-project.io,resources=serverlesses/finalizers,verbs=update

// initReconciler injects the required configuration into the declarative reconciler.
func (r *ServerlessReconciler) initReconciler(mgr ctrl.Manager) error {
	manifestResolver := &ManifestResolver{
		chartPath: r.ChartPath,
	}

	return r.Inject(mgr, &v1alpha1.Serverless{},
		declarative.WithManifestResolver(manifestResolver),
		declarative.WithResourcesReady(true),
	)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServerlessReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Config = mgr.GetConfig()
	if err := r.initReconciler(mgr); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Serverless{}).
		Complete(r)
}

// ManifestResolver represents the chart information for the passed Sample resource.
type ManifestResolver struct {
	chartPath string
}

// Get returns the chart information to be processed.
func (m *ManifestResolver) Get(obj types.BaseCustomObject, _ logr.Logger) (types.InstallationSpec, error) {
	serverless, valid := obj.(*v1alpha1.Serverless)
	if !valid {
		return types.InstallationSpec{},
			fmt.Errorf("invalid type conversion for %s", client.ObjectKeyFromObject(obj))
	}

	// default empty fields
	serverless.Spec.Default()

	flags, err := structToFlags(serverless.Spec)
	if err != nil {
		return types.InstallationSpec{},
			fmt.Errorf("resolving manifest failed: %w", err)
	}

	return types.InstallationSpec{
		ChartPath: m.chartPath,
		ChartFlags: types.ChartFlags{
			ConfigFlags: types.Flags{
				"Namespace":       chartNs,
				"CreateNamespace": true,
			},
			SetFlags: flags,
		},
	}, nil
}

func structToFlags(obj interface{}) (flags types.Flags, err error) {
	data, err := json.Marshal(obj)

	if err != nil {
		return
	}

	err = json.Unmarshal(data, &flags)
	return
}
