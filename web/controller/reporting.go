package controller

import (
	// . "eaciit/scbtwist/library/core"
	// . "eaciit/scbtwist/library/models"
	"eaciit/scbtwist/web/helper"

	"github.com/eaciit/knot/knot.v1"
)

type ReportingController struct {
	App
}

func CreateReportingController() *ReportingController {
	var controller = new(ReportingController)
	return controller
}

func (m *ReportingController) GetData(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	return helper.CreateResult(true, nil, "success")
}
