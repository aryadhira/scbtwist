package controller

import (
	// . "eaciit/scbtwist/library/core"
	// . "eaciit/scbtwist/library/models"
	"eaciit/scbtwist/web/helper"

	"github.com/eaciit/knot/knot.v1"
)

type SCMManagementController struct {
	App
}

func CreateSCMManagementController() *SCMManagementController {
	var controller = new(SCMManagementController)
	return controller
}

func (m *SCMManagementController) GetData(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	return helper.CreateResult(true, nil, "success")
}
