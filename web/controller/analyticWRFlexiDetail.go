package controller

import (
	. "eaciit/scbtwist/library/core"
	_ "eaciit/scbtwist/library/helper"
	. "eaciit/scbtwist/library/models"
	"eaciit/scbtwist/web/helper"
	c "github.com/eaciit/crowd"
	"github.com/eaciit/dbox"
	_ "github.com/eaciit/dbox/dbc/mongo"
	"github.com/eaciit/knot/knot.v1"
	"github.com/eaciit/toolkit"
	_ "math"
	"sort"
	_ "strconv"
	"strings"
	"time"
)

type AnalyticWRFlexiDetailController struct {
	App
}

func CreateAnalyticWRFlexiDetailController() *AnalyticWRFlexiDetailController {
	var controller = new(AnalyticWRFlexiDetailController)
	return controller
}

type DataItems struct {
	DirectionNo    int
	DirectionDesc  int
	WsCategoryNo   int
	WsCategoryDesc string
	Frequency      int
}

type DataItemsResult struct {
	DirectionNo    int
	DirectionDesc  int
	WsCategoryNo   int
	WsCategoryDesc string
	Hours          float64
	Contribution   float64
	Frequency      int
}

type DataItemsGroup struct {
	DirectionNo    int
	DirectionDesc  int
	WsCategoryNo   int
	WsCategoryDesc string
}

var (
	WindRoseResult []toolkit.M
	section        int
	degree         int
	fullWSCatList  []string
)

func getFullWSCategory() {
	wsCatList := []int{0, 1, 2, 3, 4}
	wsCatListDesc := []string{
		"0 to 4m/s",
		"4 to 8m/s",
		"8 to 12m/s",
		"12 to 16m/s",
		"16 to 20m/s",
		"20 and above",
	}
	fullWSCatList = []string{}
	maxSection := section
	// if degree == 45 {
	// 	maxSection = 360 / 15
	// }
	for i := 0; i < maxSection; i++ {
		for x, wscat := range wsCatList {
			fullWSCatList = append(fullWSCatList, toolkit.ToString(i)+
				"_"+toolkit.ToString(wscat)+"_"+wsCatListDesc[x])
		}
	}
}

func getWsCategory(ws float64) (int, string) {
	catNo := 0
	catDesc := "0 to 4m/s"
	if ws >= 20 {
		catNo = 5
		catDesc = "20 and above"
	} else if ws >= 16 {
		catNo = 4
		catDesc = "16 to 20m/s"
	} else if ws >= 12 {
		catNo = 3
		catDesc = "12 to 16m/s"
	} else if ws >= 8 {
		catNo = 2
		catDesc = "8 to 12m/s"
	} else if ws >= 4 {
		catNo = 1
		catDesc = "4 to 8m/s"
	}

	return catNo, catDesc
}

func getDirection(windDir float64, breakDown int) (int, int) {
	dirNo := 0

	dirDescs := make([]int, breakDown)
	direction := 0
	divider := 360 / breakDown
	for j := 0; j < breakDown; j++ {
		dirDescs[j] = direction
		direction += divider
	}
	for { /*jika wind dir negatif, tambah terus dengan 360 sampai positif*/
		if windDir < 0 {
			windDir += 360.0
		} else {
			break
		}
	}

	//dirCalc := math.Mod(windDir, 360.0) /*modulus 360 biar kalo dibagi 'divider' hasilnya <= section*/
	dirNo = int(toolkit.RoundingAuto64(toolkit.Div(windDir, float64(divider)), 0))

	if dirNo > (breakDown - 1) {
		dirNo = 0
	}
	dirNo45 := dirNo
	// if degree == 45 {
	// 	if dirNo > 0 && dirNo < 8 {
	// 		dirNo45 = dirNo * 3 /*adding section from 8 (45 degree) to 24 (15 degree)*/
	// 	}
	// }

	return dirNo45, dirDescs[dirNo]
}

func (m *AnalyticWRFlexiDetailController) GetFlexiDataEachTurbine(k *knot.WebContext) interface{} {
	k.Config.OutputType = knot.OutputJson
	WindRoseResult = []toolkit.M{}
	maxValue := 0.0
	tkMaxVal := toolkit.M{}

	type PayloadWindRose struct {
		Project   string
		Turbine   []interface{}
		DateStart time.Time
		DateEnd   time.Time
		BreakDown string
	}
	type MiniMetTower struct {
		DHubWD88mAvg float64
		VHubWS90mAvg float64
	}

	type MiniScada struct {
		NacelDirection float64
		AvgWindSpeed   float64
		Turbine        string
	}
	p := new(PayloadWindRose)
	e := k.GetPayload(&p)
	if e != nil {
		return helper.CreateResult(false, nil, e.Error())
	}

	tStart, _ := time.Parse("2006-01-02", p.DateStart.UTC().Format("2006-01-02"))
	tEnd, _ := time.Parse("2006-01-02 15:04:05", p.DateEnd.UTC().Format("2006-01-02")+" 23:59:59")
	bufferTurbine := []string{}
	for _, val := range p.Turbine {
		bufferTurbine = append(bufferTurbine, val.(string))
	}
	degree = toolkit.ToInt(p.BreakDown, toolkit.RoundingAuto)
	section = degree
	getFullWSCategory()

	coId := 0
	sort.Strings(bufferTurbine)
	turbine := []string{"MetTower"}
	turbine = append(turbine, bufferTurbine...)
	var filter []*dbox.Filter
	filter = append(filter, dbox.Ne("_id", ""))
	filter = append(filter, dbox.Gte("dateinfo.dateid", tStart))
	filter = append(filter, dbox.Lte("dateinfo.dateid", tEnd))
	filter = append(filter, dbox.In("turbine", p.Turbine...))

	scadas := []MiniScada{}

	csr, _ := DB().Connection.NewQuery().From(new(ScadaDataNew).TableName()).
		Where(dbox.And(filter...)).Cursor(nil) //.Order("turbine")

	e = csr.Fetch(&scadas, 0, false)
	if e != nil {
		return helper.CreateResult(false, nil, e.Error())
	}
	defer csr.Close()

	for _, turbineVal := range turbine {
		coId++
		groupdata := toolkit.M{}
		groupdata.Set("Index", coId)
		groupdata.Set("Name", turbineVal)

		if turbineVal != "MetTower" {
			// data := []MiniScada{}
			// if scadadata.Has(turbineVal) {
			// 	data = scadadata[turbineVal].([]MiniScada)
			// }

			data := c.From(&scadas).Where(func(x interface{}) interface{} {
				y := x.(MiniScada)

				return y.Turbine == turbineVal
			}).Exec().Result.Data().([]MiniScada)

			if toolkit.SliceLen(data) > 0 {
				totalDuration := float64(len(data)) /* Tot data * 2 for get total minutes*/
				datas := c.From(&data).Apply(func(x interface{}) interface{} {
					dt := x.(MiniScada)
					var di DataItems

					dirNo, dirDesc := getDirection(dt.NacelDirection, section)
					wsNo, wsDesc := getWsCategory(dt.AvgWindSpeed)

					di.DirectionNo = dirNo
					di.DirectionDesc = dirDesc
					di.WsCategoryNo = wsNo
					di.WsCategoryDesc = wsDesc
					di.Frequency = 1

					return di
				}).Exec().Group(func(x interface{}) interface{} {
					dt := x.(DataItems)

					var dig DataItemsGroup
					dig.DirectionNo = dt.DirectionNo
					dig.DirectionDesc = dt.DirectionDesc
					dig.WsCategoryNo = dt.WsCategoryNo
					dig.WsCategoryDesc = dt.WsCategoryDesc

					return dig
				}, nil).Exec()

				dts := datas.Apply(func(x interface{}) interface{} {
					kv := x.(c.KV)
					vv := kv.Key.(DataItemsGroup)
					vs := kv.Value.([]DataItems)

					// sumDuration := c.From(&vs).Sum(func(x interface{}) interface{} {
					// 	dt := x.(DataItems)
					// 	return dt.Hours
					// }).Exec().Result.Sum

					sumFreq := c.From(&vs).Sum(func(x interface{}) interface{} {
						dt := x.(DataItems)
						return dt.Frequency
					}).Exec().Result.Sum

					var di DataItemsResult
					// H := sumFreq
					di.DirectionNo = vv.DirectionNo
					di.DirectionDesc = vv.DirectionDesc
					di.WsCategoryNo = vv.WsCategoryNo
					di.WsCategoryDesc = vv.WsCategoryDesc
					di.Hours = toolkit.Div(sumFreq, 6.0)
					di.Contribution = toolkit.RoundingAuto64(toolkit.Div(sumFreq, totalDuration)*100.0, 2)

					key := turbineVal + "_" + toolkit.ToString(di.DirectionNo)

					if !tkMaxVal.Has(key) {
						tkMaxVal.Set(key, di.Contribution)
					} else {
						tkMaxVal.Set(key, tkMaxVal.GetFloat64(key)+di.Contribution)
					}

					di.Frequency = int(sumFreq)

					return di
				}).Exec()

				results := dts.Result.Data().([]DataItemsResult)
				wsCategoryList := []string{}
				for _, data := range results {
					wsCategoryList = append(wsCategoryList, toolkit.ToString(data.DirectionNo)+
						"_"+toolkit.ToString(data.WsCategoryNo)+"_"+data.WsCategoryDesc)
				}
				splitCatList := []string{}
				for _, wsCat := range fullWSCatList {
					if !toolkit.HasMember(wsCategoryList, wsCat) {
						splitCatList = strings.Split(wsCat, "_")
						emptyRes := DataItemsResult{}
						emptyRes.DirectionNo = toolkit.ToInt(splitCatList[0], toolkit.RoundingAuto)
						divider := section

						emptyRes.DirectionDesc = (360 / divider) * emptyRes.DirectionNo
						emptyRes.WsCategoryNo = toolkit.ToInt(splitCatList[1], toolkit.RoundingAuto)
						emptyRes.WsCategoryDesc = splitCatList[2]
						results = append(results, emptyRes)
					}
				}
				groupdata.Set("Data", results)

				WindRoseResult = append(WindRoseResult, groupdata)
			} else {
				splitCatList := []string{}
				results := []DataItemsResult{}
				for _, wsCat := range fullWSCatList {
					splitCatList = strings.Split(wsCat, "_")
					emptyRes := DataItemsResult{}
					emptyRes.DirectionNo = toolkit.ToInt(splitCatList[0], toolkit.RoundingAuto)
					divider := section

					emptyRes.DirectionDesc = (360 / divider) * emptyRes.DirectionNo
					emptyRes.WsCategoryNo = toolkit.ToInt(splitCatList[1], toolkit.RoundingAuto)
					emptyRes.WsCategoryDesc = splitCatList[2]
					results = append(results, emptyRes)
				}
				groupdata.Set("Data", results)
				WindRoseResult = append(WindRoseResult, groupdata)
			}
		} else {
			data := []MiniMetTower{}
			// data = scadadata[turbineVal].([]MiniMetTower)

			csrMet, _ := DB().Connection.NewQuery().From(new(MetTower).TableName()).
				Where(dbox.And(filter[0:3]...)).Cursor(nil)

			e = csrMet.Fetch(&data, 0, false)
			if e != nil {
				return helper.CreateResult(false, nil, e.Error())
			}
			defer csrMet.Close()

			if toolkit.SliceLen(data) > 0 {
				totalDuration := float64(len(data)) // * 10.0 /* Tot data * 2 for get total minutes*/
				datas := c.From(&data).Apply(func(x interface{}) interface{} {
					dt := x.(MiniMetTower)
					var di DataItems

					dirNo, dirDesc := getDirection(dt.DHubWD88mAvg, section)
					wsNo, wsDesc := getWsCategory(dt.VHubWS90mAvg)

					di.DirectionNo = dirNo
					di.DirectionDesc = dirDesc
					di.WsCategoryNo = wsNo
					di.WsCategoryDesc = wsDesc
					di.Frequency = 1

					return di
				}).Exec().Group(func(x interface{}) interface{} {
					dt := x.(DataItems)

					var dig DataItemsGroup
					dig.DirectionNo = dt.DirectionNo
					dig.DirectionDesc = dt.DirectionDesc
					dig.WsCategoryNo = dt.WsCategoryNo
					dig.WsCategoryDesc = dt.WsCategoryDesc

					return dig
				}, nil).Exec()

				dts := datas.Apply(func(x interface{}) interface{} {
					kv := x.(c.KV)
					vv := kv.Key.(DataItemsGroup)
					vs := kv.Value.([]DataItems)

					// sumDuration := c.From(&vs).Sum(func(x interface{}) interface{} {
					// 	dt := x.(DataItems)
					// 	return dt.Hours
					// }).Exec().Result.Sum

					sumFreq := c.From(&vs).Sum(func(x interface{}) interface{} {
						dt := x.(DataItems)
						return dt.Frequency
					}).Exec().Result.Sum

					var di DataItemsResult
					// H := sumFreq * 10.0
					di.DirectionNo = vv.DirectionNo
					di.DirectionDesc = vv.DirectionDesc
					di.WsCategoryNo = vv.WsCategoryNo
					di.WsCategoryDesc = vv.WsCategoryDesc
					di.Hours = toolkit.Div(sumFreq, 6.0)
					di.Contribution = toolkit.RoundingAuto64(toolkit.Div(sumFreq, totalDuration)*100.0, 2)

					key := turbineVal + "_" + toolkit.ToString(di.DirectionNo)
					if !tkMaxVal.Has(key) {
						tkMaxVal.Set(key, di.Contribution)
					} else {
						tkMaxVal.Set(key, tkMaxVal.GetFloat64(key)+di.Contribution)
					}

					di.Frequency = int(sumFreq)

					return di
				}).Exec()

				results := dts.Result.Data().([]DataItemsResult)
				wsCategoryList := []string{}
				for _, data := range results {
					wsCategoryList = append(wsCategoryList, toolkit.ToString(data.DirectionNo)+
						"_"+toolkit.ToString(data.WsCategoryNo)+"_"+data.WsCategoryDesc)
				}
				splitCatList := []string{}
				for _, wsCat := range fullWSCatList {
					if !toolkit.HasMember(wsCategoryList, wsCat) {
						splitCatList = strings.Split(wsCat, "_")
						emptyRes := DataItemsResult{}
						emptyRes.DirectionNo = toolkit.ToInt(splitCatList[0], toolkit.RoundingAuto)
						divider := section

						emptyRes.DirectionDesc = (360 / divider) * emptyRes.DirectionNo
						emptyRes.WsCategoryNo = toolkit.ToInt(splitCatList[1], toolkit.RoundingAuto)
						emptyRes.WsCategoryDesc = splitCatList[2]
						results = append(results, emptyRes)
					}
				}
				groupdata.Set("Data", results)

				WindRoseResult = append(WindRoseResult, groupdata)
			} else {
				splitCatList := []string{}
				results := []DataItemsResult{}
				for _, wsCat := range fullWSCatList {
					splitCatList = strings.Split(wsCat, "_")
					emptyRes := DataItemsResult{}
					emptyRes.DirectionNo = toolkit.ToInt(splitCatList[0], toolkit.RoundingAuto)
					divider := section

					emptyRes.DirectionDesc = (360 / divider) * emptyRes.DirectionNo
					emptyRes.WsCategoryNo = toolkit.ToInt(splitCatList[1], toolkit.RoundingAuto)
					emptyRes.WsCategoryDesc = splitCatList[2]
					results = append(results, emptyRes)
				}
				groupdata.Set("Data", results)
				WindRoseResult = append(WindRoseResult, groupdata)
			}
		}
	}

	for _, val := range tkMaxVal {
		if val.(float64) > maxValue {
			maxValue = val.(float64)
		}
	}

	switch {
	case maxValue >= 90 && maxValue <= 100:
		maxValue = 100
	case maxValue >= 80 && maxValue < 90:
		maxValue = 90
	case maxValue >= 70 && maxValue < 80:
		maxValue = 80
	case maxValue >= 60 && maxValue < 70:
		maxValue = 70
	case maxValue >= 50 && maxValue < 60:
		maxValue = 60
	case maxValue >= 40 && maxValue < 50:
		maxValue = 50
	case maxValue >= 30 && maxValue < 40:
		maxValue = 40
	case maxValue >= 20 && maxValue < 30:
		maxValue = 30
	case maxValue >= 10 && maxValue < 20:
		maxValue = 20
	case maxValue >= 0 && maxValue < 10:
		maxValue = 10
	}

	data := struct {
		WindRose toolkit.Ms
		MaxValue float64
	}{
		WindRose: WindRoseResult,
		MaxValue: maxValue,
	}

	return helper.CreateResult(true, data, "success")

}
