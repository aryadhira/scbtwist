package controller

import (
	// . "eaciit/scbtwist/library/core"
	// . "eaciit/scbtwist/library/models"
	"eaciit/scbtwist/web/helper"

	"github.com/eaciit/knot/knot.v1"
)

type TurbineHealthController struct {
	App
}

func CreateTurbineHealthController() *TurbineHealthController {
	var controller = new(TurbineHealthController)
	return controller
}

func (m *TurbineHealthController) GetData(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	return helper.CreateResult(true, nil, "success")
}
