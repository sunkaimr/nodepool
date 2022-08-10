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
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	poolv1 "nodepool/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// NodePoolReconciler reconciles a NodePool object
type NamespaceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=nodes.sunkai.xyz,resources=nodepools,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=nodes.sunkai.xyz,resources=nodepools/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=nodes.sunkai.xyz,resources=nodepools/finalizers,verbs=update
// Reconcile, ns发生变动，只需在创建ns时创建对应的nodepool
func (r *NamespaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	ns := corev1.Namespace{}
	err := r.Get(ctx, req.NamespacedName, &ns)
	if err != nil {
		if errors.IsNotFound(err) {
			l.Info(fmt.Sprintf("namespace:%v has been delete", req))
			return ctrl.Result{}, nil
		}
		l.Error(err, fmt.Sprintf("error on getting namespace:%v", req))
		return ctrl.Result{}, err
	}

	if InclusionExceptionNs(ns.Name) {
		l.Info(fmt.Sprintf("namespace:%v has been exclusion", ns.Name))
		return ctrl.Result{}, err
	}

	genPool := GenerateNodePoolObj(DefaultNodePoolName, ns.Name)
	pool := poolv1.NodePool{}
	exist := true
	err = r.Get(ctx, client.ObjectKeyFromObject(genPool), &pool)
	if err != nil {
		if errors.IsNotFound(err) {
			exist = false
		} else {
			l.Info(fmt.Sprintf("error on getting namespace:%v", req))
			return ctrl.Result{}, err
		}
	}

	// nodepool不存在时创建一个新的
	if !exist {
		l.Info(fmt.Sprintf("nodepool: %v/%s not exist", genPool.Namespace, genPool.Name))
		err = r.Create(ctx, genPool)
		if err != nil {
			l.Error(err, "error on create nodepool")
			return ctrl.Result{}, err
		}
		l.Info(fmt.Sprintf("nodepool: %v/%s created", genPool.Namespace, genPool.Name))

		err = r.Status().Update(ctx, genPool)
		if err != nil {
			l.Error(err, fmt.Sprintf("failed to update nodepool status"))
			return ctrl.Result{}, err
		}
	}
	/*else {
		// nodepool存在时更新nodepool，使其恢复到默认状态
		if reflect.DeepEqual(pool.Spec.NodeSelector, genPool.Spec.NodeSelector) {
			return ctrl.Result{}, nil
		}

		pool.Spec.NodeSelector = genPool.Spec.NodeSelector
		err = r.Update(ctx, &pool)
		if err != nil {
			l.Error(err, "error on update nodepool")
			return ctrl.Result{}, err
		}
		l.Info(fmt.Sprintf("nodepool: %v/%s updated", genPool.Namespace, genPool.Name))
	}
	*/
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NamespaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Namespace{}).
		Complete(r)
}
