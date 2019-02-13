package controller

import (
	"github.com/kyma-project/test-infra/development/prow-addons-ctrl-manager/pkg/controller/notifier"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, notifier.Add)
}
