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
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	poolv1 "nodepool/api/v1"
)

// NodePoolReconciler reconciles a NodePool object
type NodePoolReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=nodes.sunkai.xyz,resources=nodepools,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=nodes.sunkai.xyz,resources=nodepools/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=nodes.sunkai.xyz,resources=nodepools/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NodePool object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
// Reconcile. nodepool变动时处理逻辑
func (r *NodePoolReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	exist := true
	pool := poolv1.NodePool{}
	err := r.Get(ctx, req.NamespacedName, &pool)
	if err != nil {
		if errors.IsNotFound(err) {
			l.Info(fmt.Sprintf("nodepool: %v not exist", req.NamespacedName))
			exist = false
		} else {
			l.Error(err, fmt.Sprintf("%v", req))
			return ctrl.Result{}, err
		}
	}

	// 判断ns是否需要创建nodepool
	if InclusionExceptionNs(req.Namespace) {
		// 不需要创建nodepool，但是nodepool已经存在了就删除掉
		if exist  {
			err = r.Delete(ctx, &pool)
			if err != nil {
				l.Error(err, "error on delete nodepool")
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	if !exist {
		// 默认的nodepool被删除时自动创建
		if req.Name == DefaultNodePoolName {
			pool := GenerateNodePoolObj(DefaultNodePoolName, req.Namespace)
			err = r.Create(ctx, pool)
			if err != nil {
				l.Error(err, "error on create nodepool")
				return ctrl.Result{}, err
			}
			l.Info(fmt.Sprintf("default nodepool: %s/%s not exist and created", pool.Namespace, pool.Name))
		} else {
			// 非默认的nodepool被删除无需处理
			l.Info(fmt.Sprintf("nodepool: %s/%s not exist", pool.Namespace, pool.Name))
			return ctrl.Result{}, nil
		}
	} else {
		// nodepool 更新时恢复其spec中的默认字段
		genPool := GenerateNodePoolObj(DefaultNodePoolName, pool.Namespace)
		if !reflect.DeepEqual(pool.Spec.NodeSelector, genPool.Spec.NodeSelector) {
			pool.Spec.NodeSelector = genPool.Spec.NodeSelector
			err = r.Update(ctx, &pool)
			if err != nil {
				l.Error(err, "error on update nodepool")
				return ctrl.Result{}, err
			}
			l.Info(fmt.Sprintf("nodepool: %s/%s change and recovered", pool.Namespace, pool.Name))
		}
	}

	nodeList := corev1.NodeList{}
	err = r.List(ctx, &nodeList)
	if err != nil {
		l.Error(err, fmt.Sprintf("error on getting all nodes"))
		return ctrl.Result{}, err
	}

	needUpdate, nodes := FindMatchNodesByNodepool(&nodeList, &pool)
	if needUpdate {
		pool.Status.Nodes = nodes
		err = r.Status().Update(ctx, &pool)
		if err != nil {
			l.Error(err, fmt.Sprintf("failed to add node to nodepool:%v", pool))
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NodePoolReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&poolv1.NodePool{}).
		Complete(r)
}
