package controller

import (
	"gitlab.beopenit.com/cloud/onboarding-operator-kubernetes/pkg/controller/environment"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, environment.Add)
}
