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
	appsv1 "k8s.io/api/apps/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DeploymentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *DeploymentReconciler) Reconcile(_ context.Context, req ctrl.Request) (ctrl.Result, error) {
	var err error

	klog.V(3).Infof("reconciling deployment %s", req.String())
	deployment := &appsv1.Deployment{}
	if err = r.Get(context.TODO(), req.NamespacedName, deployment); err != nil {
		klog.Error(err)
		return ctrl.Result{RequeueAfter: requeuePeriod}, err
	}

	var toPatch bool
	if toPatch, err = containerIterator(deployment.Spec.Template.Spec.InitContainers); err != nil {
		klog.Error(err)
		return ctrl.Result{RequeueAfter: requeuePeriod}, err
	}
	if toPatch, err = containerIterator(deployment.Spec.Template.Spec.Containers); err != nil {
		klog.Error(err)
		return ctrl.Result{RequeueAfter: requeuePeriod}, err
	}

	if toPatch {
		if err := r.Update(context.TODO(), deployment); err != nil {
			if kerrors.IsConflict(err) {
				klog.V(3).Info(err)
			} else {
				klog.Error(err)
			}
			return ctrl.Result{RequeueAfter: requeuePeriod}, err
		}
		klog.Infof("deployment %s patched", req.String())
	}

	klog.V(3).Infof("deployment %s reconciled", req.String())
	return ctrl.Result{}, nil
}

func (r *DeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}).
		WithEventFilter(commonPredicate()).
		Complete(r)
}
