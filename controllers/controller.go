package controllers

import (
	ctrl "sigs.k8s.io/controller-runtime"
)

var ExceptionNs = make([]string, 1)

func NameSpaceControllerRun(mgr ctrl.Manager)  {
	if err := (&NamespaceReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		ctrl.Log.Error(err, "unable to create controller", "controller", "namespace")
		panic(err)
	}
}


func NodePoolControllerRun(mgr ctrl.Manager)  {
	if err := (&NodePoolReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		ctrl.Log.Error(err, "unable to create controller", "controller", "nodepool")
		panic(err)
	}
}

func NodeControllerRun(mgr ctrl.Manager)  {
	if err := (&NodeReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		ctrl.Log.Error(err, "unable to create controller", "controller", "node")
		panic(err)
	}
}

func InclusionExceptionNs(ns string)bool{
	for _, n := range ExceptionNs {
		if n == ns {
			return true
		}
	}
	return false
}