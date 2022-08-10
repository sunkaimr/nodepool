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
type NodeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=nodes.sunkai.xyz,resources=nodepools,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=nodes.sunkai.xyz,resources=nodepools/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=nodes.sunkai.xyz,resources=nodepools/finalizers,verbs=update
// Reconcile, node发生变动。增、删、改
func (r *NodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	nodeExist := true
	node := corev1.Node{}
	err := r.Get(ctx, req.NamespacedName, &node)
	if err != nil {
		if errors.IsNotFound(err) {
			l.Info(fmt.Sprintf("node:%v not exist", req))
			nodeExist = false
		} else {
			l.Error(err, fmt.Sprintf("error on getting node:%v", req))
			return ctrl.Result{}, err
		}
	}

	poolList := poolv1.NodePoolList{}
	err = r.List(ctx, &poolList)
	if err != nil {
		l.Error(err, fmt.Sprintf("error on getting all nodePool"))
		return ctrl.Result{}, err
	}

	// node被删除
	if !nodeExist {
		// 找到node所属的nodepool
		pool := FindNodepoolByNodeName(req.Name, &poolList)
		if pool == nil {
			// node没有加入任何nodepool
			return ctrl.Result{}, nil
		}

		// 删除nodepool中的node
		pool.Status.Nodes = deleteNodeFromPoolnodes(req.Name, pool.Status.Nodes)
		err = r.Status().Update(ctx, pool)
		if err != nil {
			l.Error(err, fmt.Sprintf("failed to delete node from nodepool:%v", pool))
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// node增加、修改
	found := false
	pool := FindNodepoolByNodeObj(&node, &poolList)
	if pool == nil {
		l.Info(fmt.Sprintf("node: %v not match any nodepool", node.Name))
	}

	if found {
		needUpdate := false
		needUpdate, pool.Status.Nodes = AddNodeUnique(pool.Status.Nodes, node.Name)
		if needUpdate {
			err = r.Status().Update(ctx, pool)
			if err != nil {
				l.Error(err, fmt.Sprintf("failed to add node: %v to nodepool:%v/%v",
					node.Name, pool.Namespace, pool.Name))
				return ctrl.Result{}, err
			}
			l.Info(fmt.Sprintf("add node: %v to nodepool: %v/%v",
				node.Name, pool.Namespace, pool.Name))
		}
		return ctrl.Result{}, nil
	}

	// 不知道node属于哪个nodepool，比如node的lables被移除
	nodeList := corev1.NodeList{}
	err = r.List(ctx, &nodeList)
	if err != nil {
		l.Error(err, fmt.Sprintf("error on getting all node"))
		return ctrl.Result{}, err
	}

	for _, pool := range poolList.Items {
		neeedUpdate, nodes := FindMatchNodesByNodepool(&nodeList, &pool)
		if neeedUpdate {
			pool.Status.Nodes = nodes
			err = r.Status().Update(ctx, &pool)
			if err != nil {
				l.Error(err, fmt.Sprintf("failed to add node to nodepool:%v", pool))
				return ctrl.Result{}, err
			}
			l.Info(fmt.Sprintf("update nodepool:%s/%s, nodes: %v", pool.Namespace, pool.Name, nodes))
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Node{}).
		Complete(r)
}
