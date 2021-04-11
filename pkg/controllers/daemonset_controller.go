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
	kerrors "k8s.io/apimachinery/pkg/api/errors"
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

func (r *DaemonsetReconciler) Reconcile(_ context.Context, req ctrl.Request) (ctrl.Result, error) {
	var err error

	klog.V(3).Infof("reconciling daemonset %s", req.String())
	daemonset := &appsv1.DaemonSet{}
	if err = r.Get(context.TODO(), req.NamespacedName, daemonset); err != nil {
		klog.Error(err)
		return ctrl.Result{RequeueAfter: requeuePeriod}, err
	}

	var toPatch bool
	if toPatch, err = containerIterator(daemonset.Spec.Template.Spec.InitContainers); err != nil {
		klog.Error(err)
		return ctrl.Result{RequeueAfter: requeuePeriod}, err
	}
	if toPatch, err = containerIterator(daemonset.Spec.Template.Spec.Containers); err != nil {
		klog.Error(err)
		return ctrl.Result{RequeueAfter: requeuePeriod}, err
	}

	if toPatch {
		if err := r.Update(context.TODO(), daemonset); err != nil {
			if kerrors.IsConflict(err) {
				klog.V(3).Info(err)
			} else {
				klog.Error(err)
			}
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
