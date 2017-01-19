package controller

import (
	// . "eaciit/scbtwist/library/core"
	// . "eaciit/scbtwist/library/models"
	"eaciit/scbtwist/web/helper"

	"github.com/eaciit/knot/knot.v1"
)

type DiyViewController struct {
	App
}

func CreateDiyViewController() *DiyViewController {
	var controller = new(DiyViewController)
	return controller
}

func (m *DiyViewController) GetData(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	return helper.CreateResult(true, nil, "success")
}
