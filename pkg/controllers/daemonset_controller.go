/*


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
	"image-clone-controller/pkg/imagesManagement"
	"k8s.io/klog/v2"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DaemonsetReconciler reconciles a Deployment object
type DaemonsetReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=apps,resources=deployments,daemonsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,daemonsets/status,verbs=get;update;patch

func (r *DaemonsetReconciler) Reconcile(_ context.Context, req ctrl.Request) (ctrl.Result, error) {
	var err error

	klog.V(3).Infof("reconciling daemonset %s", req.String())
	daemonset := &appsv1.DaemonSet{}
	if err = r.Get(context.TODO(), req.NamespacedName, daemonset); err != nil {
		klog.Error(err)
		return ctrl.Result{RequeueAfter: requeuePeriod}, err
	}

	var toPatch bool
	for _, c := range daemonset.Spec.Template.Spec.Containers {
		backupImageName, err := imagesManagement.Get().EnforceBackup(c.Image)
		if err != nil {
			klog.Error(err)
			return ctrl.Result{RequeueAfter: requeuePeriod}, err
		}
		if backupImageName != c.Image {
			c.Image = backupImageName
			toPatch = true
		}
	}

	if toPatch {
		if err := r.Update(context.TODO(), daemonset); err != nil {
			klog.Error(err)
			return ctrl.Result{RequeueAfter: requeuePeriod}, err
		}
	}

	klog.V(3).Infof("daemonset %s reconciled", req.String())
	return ctrl.Result{}, nil
}

func (r *DaemonsetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.DaemonSet{}).
		WithEventFilter(commonPredicate()).
		Complete(r)
}
