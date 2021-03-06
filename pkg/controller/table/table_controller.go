/*
Copyright 2019 Replicated, Inc.

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

package table

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	schemasv1alpha3 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha3"
	"github.com/schemahero/schemahero/pkg/logger"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Add creates a new Table Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileTable{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("table-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Table
	err = c.Watch(&source.Kind{Type: &schemasv1alpha3.Table{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Add an informer on pods, which are created to deploy schemas. the informer will
	// update the status of the table custom resource and do a little garbage collection
	generatedClient := kubernetes.NewForConfigOrDie(mgr.GetConfig())
	generatedInformers := kubeinformers.NewSharedInformerFactory(generatedClient, time.Minute)
	err = mgr.Add(manager.RunnableFunc(func(s <-chan struct{}) error {
		generatedInformers.Start(s)
		<-s
		return nil
	}))
	if err != nil {
		return err
	}

	// watch for pods because pods are how schemahero deploys, and the lifecycle of a pod is important
	// to ensure that we are deployed
	err = c.Watch(&source.Informer{
		Informer: generatedInformers.Core().V1().Pods().Informer(),
	}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileTable{}

// ReconcileTable reconciles a Table object
type ReconcileTable struct {
	client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Table object and makes changes based on the state read
// and what is in the Table.Spec
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=schemas.schemahero.io,resources=tables,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=schemas.schemahero.io,resources=tables/status,verbs=get;update;patch
func (r *ReconcileTable) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	instance, instanceErr := r.getInstance(request)
	if instanceErr == nil {
		fmt.Printf("instance\n")
		result, err := r.reconcileInstance(instance)
		if err != nil {
			logger.Error(err)
		}
		return result, err
	}

	pod := &corev1.Pod{}
	podErr := r.Get(context.Background(), request.NamespacedName, pod)
	if podErr == nil {
		result, err := r.reconcilePod(pod)
		if err != nil {
			logger.Error(err)
		}
		return result, err
	}

	return reconcile.Result{}, errors.New("unknown error in table reconciler")
}
