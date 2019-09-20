package rpc

import (
	"github.com/pkg/errors"
	"log"
	"net/http"

	erpc "github.com/Varunram/essentials/rpc"
	utils "github.com/Varunram/essentials/utils"
	openx "github.com/YaleOpenLab/openx/database"
	core "github.com/test/blah/core"
)

func setupAdminHandlers() {
	flagProject()
}

var AdminRPC = map[int][]string{
	1: []string{"/admin/flag", "projIndex"}, // GET
}

func adminValidateHelper(w http.ResponseWriter, r *http.Request) (openx.User, error) {
	var user openx.User

	username := r.URL.Query()["username"][0]
	token := r.URL.Query()["token"][0]

	user, err := core.ValidateUser(username, token)
	if err != nil {
		log.Println(err)
		erpc.ResponseHandler(w, erpc.StatusBadRequest)
		return user, err
	}

	if !user.Admin {
		erpc.ResponseHandler(w, erpc.StatusUnauthorized)
		return user, errors.New("unauthorized")
	}

	return user, nil
}

func flagProject() {
	http.HandleFunc(AdminRPC[1][0], func(w http.ResponseWriter, r *http.Request) {
		err := erpc.CheckGet(w, r)
		if err != nil {
			log.Println(err)
			return
		}

		err = checkReqdParams(w, r, AdminRPC[1][1:])
		if err != nil {
			return
		}

		user, err := adminValidateHelper(w, r)
		if err != nil {
			log.Println(err)
			erpc.ResponseHandler(w, erpc.StatusUnauthorized)
		}

		projIndex, err := utils.ToInt(r.URL.Query()["projIndex"][0])
		if err != nil {
			log.Println(err)
			erpc.ResponseHandler(w, erpc.StatusBadRequest)
			return
		}

		err = core.MarkFlagged(projIndex, user.Index)
		if err != nil {
			log.Println(err)
			erpc.ResponseHandler(w, erpc.StatusInternalServerError)
			return
		}

		erpc.ResponseHandler(w, erpc.StatusOK)
	})
}
