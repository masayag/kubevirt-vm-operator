package fedora

import (
	"context"
	"fmt"
	"reflect"

	guestv1alpha1 "github.com/masayag/kubevirt-vm-operator/pkg/apis/guest/v1alpha1"
	"github.com/masayag/kubevirt-vm-operator/pkg/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubevirt "kubevirt.io/kubevirt/pkg/api/v1"
	"kubevirt.io/kubevirt/pkg/kubecli"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_fedora")

// Add creates a new Fedora Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	kubeClient, err := kubecli.GetKubevirtClientFromRESTConfig(mgr.GetConfig())
	if err != nil {
		log.Error(err, "Unable to get KubeVirt client")
		panic("Controller cannot operate without KubeVirt")
	}
	return &ReconcileFedora{client: mgr.GetClient(), scheme: mgr.GetScheme(), kubeClient: kubeClient}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("fedora-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Fedora
	err = c.Watch(&source.Kind{Type: &guestv1alpha1.Fedora{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileFedora implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileFedora{}

// ReconcileFedora reconciles a Fedora object
type ReconcileFedora struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client     client.Client
	scheme     *runtime.Scheme
	kubeClient kubecli.KubevirtClient
}

// Reconcile reads that state of the cluster for a Fedora object and makes changes based on the state read
// and what is in the Fedora.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileFedora) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Fedora")

	// Fetch the Fedora instance
	instance := &guestv1alpha1.Fedora{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue

			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Define a new VirtualMachine object
	vm := newVirtualMachineForCR(instance)

	/**
	TODO:

	With setting the Fedora CR as the owner of the VM, it will become responsible
	for garbage collection of the VM.
	An example of the ownerReferences from the VM:

	ownerReferences:
	- apiVersion: guest.kubevirt.io/v1alpha1
		blockOwnerDeletion: true
		controller: true
		kind: Fedora
		name: example-fedora
		uid: 5fff462e-90c3-11e9-a43c-5254007d353d

	- Should we manage VM remove using this method or by explicitly removal of the VM managed by this CR ?
	- Any cleanup needed by finalizer ?
	- How should we handle PVC / storage related entities ?

	*/
	// Set Fedora instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, vm, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this VirtualMachine already exists
	getOptions := &metav1.GetOptions{}
	existingVM, err := r.kubeClient.VirtualMachine(request.Namespace).Get(vm.Name, getOptions)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new VM")
		_, err = r.kubeClient.VirtualMachine(request.Namespace).Create(vm)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Updates CR status with new vm
		err = r.updateFedoraStatus(instance)
		if err != nil {
			reqLogger.Error(err, "Failed to update Fedora CR status")
			return reconcile.Result{}, err
		}

		// VM created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Update VirtualMachine as a result of change in CR spec
	if vmChanged(existingVM, vm) {
		existingVM.Spec.Template.Spec.Domain = vm.Spec.Template.Spec.Domain
		existingVM, err := r.kubeClient.VirtualMachine(request.Namespace).Update(existingVM)
		if err != nil {
			reqLogger.Error(err, "Failed to update VirtualMachine", "VirtualMachine.Namespace",
				existingVM.Namespace, "VirtualMachine.Name", existingVM.Name)
			return reconcile.Result{}, err
		}

		reqLogger.Info("Updated VirtualMachine", "VirtualMachine.Namespace", existingVM.Namespace,
			"VirtualMachine.Name", existingVM.Name)
		return reconcile.Result{}, nil
	}

	// VirtualMachine already exists - don't requeue
	reqLogger.Info("Skip reconcile: VirtualMachine already exists", "VirtualMachine.Namespace",
		existingVM.Namespace, "VirtualMachine.Name", existingVM.Name)
	return reconcile.Result{}, nil
}

func (r *ReconcileFedora) updateFedoraStatus(cr *guestv1alpha1.Fedora) error {
	// Update Fedora CR status with list of labelled VMs
	labelSelector := fmt.Sprintf("%s=%s", utils.CreatedByLabel, cr.Name)
	vms, err := r.kubeClient.VirtualMachine(cr.Namespace).List(&metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		log.Error(err, "Unable to get list of VMs, Status will not be updated")
	} else {
		vmNames := make([]string, len(vms.Items))
		for i := 0; i < len(vms.Items); i++ {
			vmNames[i] = vms.Items[i].Name
		}

		cr.Status.VMs = vmNames
		return r.client.Status().Update(context.TODO(), cr)
	}

	return nil
}

// TODO: Compare only fields supported by the CR
func vmChanged(srcVM *kubevirt.VirtualMachine, dstVM *kubevirt.VirtualMachine) bool {
	return !reflect.DeepEqual(srcVM.Spec.Template.Spec.Domain, dstVM.Spec.Template.Spec.Domain)
}

// returns a VirtualMachine with the same properties as specified in the cr
func newVirtualMachineForCR(cr *guestv1alpha1.Fedora) *kubevirt.VirtualMachine {
	vm := utils.GetVirtualMachine(cr)

	image := utils.GetFedoraImage(cr.Spec.OSVersion)
	//TODO: Replace with DataVolume
	utils.AddContainerDisk(vm, image)
	if len(cr.Spec.CloudInit) > 0 {
		utils.AddNoCloudDiskWitUserData(vm, cr.Spec.CloudInit)
	}

	//TODO: support other networks
	utils.AddVmPodNetwork(vm)
	return vm
}
