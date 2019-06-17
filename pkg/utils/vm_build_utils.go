package utils

import (
	guestv1alpha1 "github.com/masayag/kubevirt-vm-operator/pkg/apis/guest/v1alpha1"
	k8sv1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "kubevirt.io/kubevirt/pkg/api/v1"
)

const (
	busVirtio = "virtio"
)

const (
	// CreatedByLabel label is used as a marker for VMs that are managed by Fedora CR
	CreatedByLabel = "apis.guest.fedora/created-by"
)

// GetVirtualMachine creates a VirtualMachine instance based on the given spec provided by the CR.
// The VirtualMachine will be created under the given namespace with the provided name and specification.
// It returns a reference to the created VirtualMachine.
func GetVirtualMachine(cr *guestv1alpha1.Fedora) *v1.VirtualMachine {
	name := cr.Spec.VMName
	namespace := cr.Namespace

	labels := map[string]string{
		"kubevirt.io/vm": name,
		CreatedByLabel:   cr.Name,
	}
	running := true
	vm := v1.VirtualMachine{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1.GroupVersion.String(),
			Kind:       "VirtualMachine",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: v1.VirtualMachineSpec{
			Running: &running,
			Template: &v1.VirtualMachineInstanceTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
					Labels:    labels,
				},
				Spec: v1.VirtualMachineInstanceSpec{
					Domain: v1.DomainSpec{
						Resources: v1.ResourceRequirements{
							Requests: k8sv1.ResourceList{
								k8sv1.ResourceMemory: resource.MustParse(cr.Spec.Memory),
							},
						},
						CPU: &v1.CPU{
							Cores: cr.Spec.CPUCores,
						},
					},
				},
			},
		},
	}

	return &vm
}

// AddContainerDisk extends virtual machine with disk and volume based on a given image
// TODO: This should be replaced with DataVolume
func AddContainerDisk(vm *v1.VirtualMachine, image string) {
	spec := &vm.Spec.Template.Spec
	spec.Domain.Devices = v1.Devices{
		Disks: []v1.Disk{
			{
				Name: "containerdisk",
				DiskDevice: v1.DiskDevice{
					Disk: &v1.DiskTarget{
						Bus: busVirtio,
					},
				},
			},
		},
	}
	spec.Volumes = []v1.Volume{
		{
			Name: "containerdisk",
			VolumeSource: v1.VolumeSource{
				ContainerDisk: &v1.ContainerDiskSource{
					Image: image,
				},
			},
		},
	}
}

// AddNoCloudDiskWitUserData adds Cloud-Init data source to a given virtual machine
func AddNoCloudDiskWitUserData(vm *v1.VirtualMachine, data string) {
	spec := &vm.Spec.Template.Spec
	spec.Domain.Devices.Disks = append(spec.Domain.Devices.Disks, v1.Disk{
		Name: "cloudinitdisk",
		DiskDevice: v1.DiskDevice{
			Disk: &v1.DiskTarget{
				Bus: busVirtio,
			},
		},
	})

	spec.Volumes = append(spec.Volumes, v1.Volume{
		Name: "cloudinitdisk",
		VolumeSource: v1.VolumeSource{
			CloudInitNoCloud: &v1.CloudInitNoCloudSource{
				UserData: data,
			},
		},
	})
}

// AddVmPodNetwork adds a default pod network to the given vm
func AddVmPodNetwork(vm *v1.VirtualMachine) {
	spec := &vm.Spec.Template.Spec
	spec.Networks = []v1.Network{*v1.DefaultPodNetwork()}
	spec.Domain.Devices.Interfaces = []v1.Interface{{Name: "default",
		InterfaceBindingMethod: v1.InterfaceBindingMethod{Bridge: &v1.InterfaceBridge{}}}}
}
