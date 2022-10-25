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
	"context"

	foov1alpha1 "github.com/entgigi/foo-operator.git/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// FooReconciler reconciles a Foo object
type FooReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// RBAC permissions to monitor foo custom resources
//+kubebuilder:rbac:groups=foo.entgigi.entando.org,resources=foos,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=foo.entgigi.entando.org,resources=foos/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=foo.entgigi.entando.org,resources=foos/finalizers,verbs=update
// RBAC permissions to monitor pods
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *FooReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("reconciling Foo custom resource")

	var fooCr foov1alpha1.Foo
	// Get the Foo resource that triggered the reconciliation request
	err := r.Get(ctx, req.NamespacedName, &fooCr)
	if err != nil {
		log.Error(err, "unable to fetch Foo")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var podList corev1.PodList
	var friendFound bool = false
	err = r.List(ctx, &podList)

	if err != nil {
		log.Error(err, "unable to list pods")

	} else {
		for _, item := range podList.Items {
			if item.GetName() == fooCr.Spec.Name {
				friendFound = true
				log.Info("pod linked to a foo custom resource found", "name", item.GetName())
			}
		}

	}

	fooCr.Status.Happy = friendFound

	err = r.Status().Update(ctx, &fooCr)
	if err != nil {
		log.Error(err, "Unable to update foo's happy status", "status", friendFound)
		return ctrl.Result{}, err
	}

	log.Info("foo's happy status updated", "status", friendFound)

	log.Info("Foo custom resource reconciled")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *FooReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&foov1alpha1.Foo{}).
		Watches(&source.Kind{Type: &corev1.Pod{}}, handler.EnqueueRequestsFromMapFunc(r.mapPodsReqToFooReq)).
		Complete(r)
}

// recupera le CR Foo di eventuali pod creati modificati o cancellati
func (r *FooReconciler) mapPodsReqToFooReq(obj client.Object) []reconcile.Request {
	ctx := context.Background()
	log := log.FromContext(ctx)

	var req []reconcile.Request

	// List all the Foo custom resource
	var fooListCrd foov1alpha1.FooList
	err := r.List(context.TODO(), &fooListCrd)
	if err != nil {
		log.Error(err, "unable to list foo custom resources")
	} else {
		//
		for _, item := range fooListCrd.Items {
			if item.Spec.Name == obj.GetName() {
				req = append(req, reconcile.Request{NamespacedName: types.NamespacedName{}})
				log.Info("pod linked to a foo custom resource issued an event", "name", obj.GetName(), "namespace", obj.GetNamespace())
			}

		}
	}

	return req
}
