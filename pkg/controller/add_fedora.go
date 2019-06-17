package controller

import (
	"github.com/masayag/kubevirt-vm-operator/pkg/controller/fedora"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, fedora.Add)
}
