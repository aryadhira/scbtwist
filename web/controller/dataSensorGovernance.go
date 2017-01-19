package controller

import (
	// . "eaciit/scbtwist/library/core"
	// . "eaciit/scbtwist/library/models"
	"eaciit/scbtwist/web/helper"

	"github.com/eaciit/knot/knot.v1"
)

type DataSensorGovernanceController struct {
	App
}

func CreateDataSensorGovernanceController() *DataSensorGovernanceController {
	var controller = new(DataSensorGovernanceController)
	return controller
}

func (m *DataSensorGovernanceController) GetData(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	return helper.CreateResult(true, nil, "success")
}
