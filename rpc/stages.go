package rpc

import (
	"log"
	"net/http"

	erpc "github.com/Varunram/essentials/rpc"
	utils "github.com/Varunram/essentials/utils"

	core "github.com/test/blah/core"
)

// setupStagesHandlers sets up all stage related handlers
func setupStagesHandlers() {
	returnAllStages()
	returnSpecificStage()
	promoteStage()
}

var StagesRPC = map[int][]string{
	1: []string{"/stages/all"},
	2: []string{"/stages", "index"},
	3: []string{"/stages/promote"},
}

// returnAllStages returns all the defined stages for this specific platform.  Opensolar
// has 9 stages defined in stages.go
func returnAllStages() {
	http.HandleFunc(StagesRPC[1][0], func(w http.ResponseWriter, r *http.Request) {
		err := erpc.CheckGet(w, r)
		if err != nil {
			log.Println(err)
			return
		}

		var arr []core.Stage
		arr = append(arr, core.Stage0, core.Stage1, core.Stage2, core.Stage3, core.Stage4,
			core.Stage5, core.Stage6, core.Stage7, core.Stage8, core.Stage9)

		erpc.MarshalSend(w, arr)
	})
}

// returnSpecificStage returns details on a specific stage defined in the opensolar platform
func returnSpecificStage() {
	http.HandleFunc(StagesRPC[2][0], func(w http.ResponseWriter, r *http.Request) {
		err := erpc.CheckGet(w, r)
		if err != nil {
			log.Println(err)
			return
		}

		err = checkReqdParams(w, r, StagesRPC[2][1:])
		if err != nil {
			log.Println(err)
			return
		}

		indexx := r.URL.Query()["index"][0]

		index, err := utils.ToInt(indexx)
		if err != nil {
			log.Println("Passed index not an integer, quitting!")
			erpc.ResponseHandler(w, erpc.StatusBadRequest)
			return
		}

		var x core.Stage
		switch index {
		case 1:
			x = core.Stage1
		case 2:
			x = core.Stage2
		case 3:
			x = core.Stage3
		case 4:
			x = core.Stage4
		case 5:
			x = core.Stage5
		case 6:
			x = core.Stage6
		case 7:
			x = core.Stage7
		case 8:
			x = core.Stage8
		case 9:
			x = core.Stage9
		default:
			// default is stage0, so we don't have a case defined for it above
			x = core.Stage0
		}

		erpc.MarshalSend(w, x)
	})
}

// promoteStage returns details on a specific stage defined in the opensolar platform
func promoteStage() {
	http.HandleFunc(StagesRPC[3][0], func(w http.ResponseWriter, r *http.Request) {
		err := erpc.CheckGet(w, r)
		if err != nil {
			log.Println(err)
			return
		}

		err = checkReqdParams(w, r, StagesRPC[2][1:])
		if err != nil {
			log.Println(err)
			return
		}

		indexx := r.URL.Query()["index"][0]

		index, err := utils.ToInt(indexx)
		if err != nil {
			log.Println("Passed index not an integer, quitting!")
			erpc.ResponseHandler(w, erpc.StatusBadRequest)
			return
		}

		err = core.StageXtoY(index)
		if err != nil {
			log.Println(err)
			erpc.ResponseHandler(w, erpc.StatusInternalServerError)
			return
		}
		erpc.ResponseHandler(w, erpc.StatusOK)
	})
}
