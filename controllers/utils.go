package controllers

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	poolv1 "nodepool/api/v1"
	"sort"
)

const (
	DefaultNodePoolName = "default"
	LableNodePoolKey    = "nodepool"
)

// GenerateNodePoolObj Generate NodePool object
func GenerateNodePoolObj(name, namespace string) *poolv1.NodePool {
	return &poolv1.NodePool{
		TypeMeta: metav1.TypeMeta{
			Kind:       "NodePool",
			APIVersion: "nodes.sunkai.xyz/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: poolv1.NodePoolSpec{
			NodeSelector: map[string]string{LableNodePoolKey: namespace},
		},
	}
}

func AddNodeUnique(nodes []string, str ...string) (changed bool, newNodes []string) {
	if len(str) == 0 {
		return false, nodes
	}

	if len(nodes) == 0 {
		newNodes = append(nodes, str...)
		sort.Strings(newNodes)
		return true, newNodes
	}

	for _, s := range str {
		found := false
		for _, node := range nodes {
			if s == node {
				found = true
				break
			}
		}
		if !found {
			newNodes = append(newNodes, s)
			changed = true
		}
	}

	newNodes = append(nodes, newNodes...)
	sort.Strings(newNodes)
	return changed, newNodes
}

func FindMatchNodesByNodepool(allNodes *corev1.NodeList, pool *poolv1.NodePool) (bool, []string) {
	haveNodePoolLables := make([]string, 0)
	// 找出符合nodepool的node
	for i := 0; i < len(allNodes.Items); i++ {
		node := &allNodes.Items[i]
		if nodeLabVal, ok := node.Labels[LableNodePoolKey]; !ok {
			continue
		} else if nodeLabVal == pool.Spec.NodeSelector[LableNodePoolKey] {
			haveNodePoolLables = append(haveNodePoolLables, node.Name)
		}
	}
	sort.Strings(haveNodePoolLables)
	if len(haveNodePoolLables) != len(pool.Status.Nodes) {
		return true, haveNodePoolLables
	}

	// 判断pool.status.node是否需要改变
	changed := false
	for i := 0; i < len(haveNodePoolLables); i++ {
		if haveNodePoolLables[i] != pool.Status.Nodes[i] {
			changed = true
			break
		}
	}

	return changed, haveNodePoolLables
}

func FindNodepoolByNodeName(node string, pools *poolv1.NodePoolList) *poolv1.NodePool {
	for _, pool := range pools.Items{
		for i, nodeName := range pool.Status.Nodes {
			if node == nodeName {
				return &pools.Items[i]
			}
		}
	}
	return nil
}

func FindNodepoolByNodeObj(node *corev1.Node, pools *poolv1.NodePoolList) *poolv1.NodePool {
	for i := 0; i < len(pools.Items); i++ {
		pool := &pools.Items[i]
		if nodeValue, ok := node.Labels[LableNodePoolKey]; ok {
			if poolValue, ok := pool.Spec.NodeSelector[LableNodePoolKey]; ok && nodeValue == poolValue {
				return pool
			}
		}
	}
	return nil
}

func deleteNodeFromPoolnodes(del string, nodes []string)(new []string){
	for _, node := range nodes {
		if del == node {
			continue
		}
		new = append(new, node)
	}
	return new
}
