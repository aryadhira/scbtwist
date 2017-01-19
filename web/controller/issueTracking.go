package controller

import (
	// . "eaciit/scbtwist/library/core"
	// . "eaciit/scbtwist/library/models"
	"eaciit/scbtwist/web/helper"

	"github.com/eaciit/knot/knot.v1"
)

type IssueTrackingController struct {
	App
}

func CreateIssueTrackingController() *IssueTrackingController {
	var controller = new(IssueTrackingController)
	return controller
}

func (m *IssueTrackingController) GetData(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	return helper.CreateResult(true, nil, "success")
}
