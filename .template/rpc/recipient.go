package rpc

import (
	"errors"
	"log"
	"net/http"

	erpc "github.com/Varunram/essentials/rpc"
	utils "github.com/Varunram/essentials/utils"
	xlm "github.com/YaleOpenLab/openx/chains/xlm"
	wallet "github.com/YaleOpenLab/openx/chains/xlm/wallet"

	core "github.com/org/plat/core"
)

// setupRecipientRPCs sets up all RPCs related to the recipient
func setupRecipientRPCs() {
	registerRecipient()
	validateRecipient()
	getAllRecipients()
	payback()
	chooseBlindAuction()
	chooseVickreyAuction()
	chooseTimeAuction()
	unlockOpenSolar()
	addEmail()
	finalizeProject()
	originateProject()
	calculateTrustLimit()
	setOneTimeUnlock()
}

// RecpRPC is a collection of all recipient RPC endpoints and their required params
var RecpRPC = map[int][]string{
	1:  []string{"/recipient/all"},                                                    // GET
	2:  []string{"/recipient/register"},                                               // POST
	3:  []string{"/recipient/validate"},                                               // GET
	4:  []string{"/recipient/payback", "assetName", "amount", "seedpwd", "projIndex"}, // POST
	8:  []string{"/recipient/auction/choose/blind"},                                   // GET
	9:  []string{"/recipient/auction/choose/vickrey"},                                 // GET
	10: []string{"/recipient/auction/choose/time"},                                    // GET
	11: []string{"/recipient/unlock/opensolar", "seedpwd", "projIndex"},               // POST
	12: []string{"/recipient/addemail", "email"},                                      // POST
	13: []string{"/recipient/finalize", "projIndex"},                                  // POST
	14: []string{"/recipient/originate", "projIndex"},                                 // POST
	15: []string{"/recipient/trustlimit", "assetName"},                                // GET
	17: []string{"/recipient/onetimeunlock", "projIndex", "seedpwd"},                  // POST
}

// recpValidateHelper is a helper that helps validates recipients in routes
func recpValidateHelper(w http.ResponseWriter, r *http.Request, options []string) (core.Recipient, error) {
	var prepRecipient core.Recipient
	var err error

	err = checkReqdParams(w, r, options)
	if err != nil {
		log.Println(err)
		erpc.ResponseHandler(w, erpc.StatusUnauthorized)
		return prepRecipient, errors.New("reqd params not present can't be empty")
	}

	var username, token string

	if r.Method == "GET" {
		username = r.URL.Query()["username"][0]
		token = r.URL.Query()["token"][0]
	} else if r.Method == "POST" {
		username = r.FormValue("username")
		token = r.FormValue("token")
	}
	prepRecipient, err = core.ValidateRecipient(username, token)
	if err != nil {
		erpc.ResponseHandler(w, erpc.StatusUnauthorized)
		log.Println("did not validate recipient", err)
		return prepRecipient, err
	}

	return prepRecipient, nil
}

// getAllRecipients gets a list of all the recipients who have registered on the platform
func getAllRecipients() {
	http.HandleFunc(RecpRPC[1][0], func(w http.ResponseWriter, r *http.Request) {
		err := erpc.CheckGet(w, r)
		if err != nil {
			log.Println(err)
			return
		}

		_, err = recpValidateHelper(w, r, RecpRPC[1][1:])
		if err != nil {
			return
		}
		recipients, err := core.RetrieveAllRecipients()
		if err != nil {
			log.Println("did not retrieve all recipients", err)
			erpc.ResponseHandler(w, erpc.StatusInternalServerError)
			return
		}
		erpc.MarshalSend(w, recipients)
	})
}

// registerRecipient creates and stores a new recipient on the platform
func registerRecipient() {
	http.HandleFunc(RecpRPC[2][0], func(w http.ResponseWriter, r *http.Request) {
		err := erpc.CheckPost(w, r)
		if err != nil {
			log.Println(err)
			return
		}

		err = checkReqdParams(w, r, RecpRPC[2][1:])
		if err != nil {
			log.Println(err)
			return
		}

		name := r.FormValue("name")
		username := r.FormValue("username")
		pwhash := r.FormValue("pwhash")
		seedpwd := r.FormValue("seedpwd")

		// check for username collision here. If the username already exists, fetch details from that and register as investor
		if core.CheckUsernameCollision(username) {
			// user already exists on the platform, need to retrieve the user
			user, err := userValidateHelper(w, r, nil) // check whether this person is a user and has params
			if err != nil {
				return
			}

			// this is the same user who wants to register as an investor now, check if encrypted seed decrypts
			seed, err := wallet.DecryptSeed(user.StellarWallet.EncryptedSeed, seedpwd)
			if err != nil {
				erpc.ResponseHandler(w, erpc.StatusInternalServerError)
				return
			}
			pubkey, err := wallet.ReturnPubkey(seed)
			if err != nil {
				erpc.ResponseHandler(w, erpc.StatusInternalServerError)
				return
			}
			if pubkey != user.StellarWallet.PublicKey {
				erpc.ResponseHandler(w, erpc.StatusUnauthorized)
				return
			}
			var a core.Recipient
			a.U = &user
			err = a.Save()
			if err != nil {
				erpc.ResponseHandler(w, erpc.StatusInternalServerError)
				return
			}
			erpc.MarshalSend(w, a)
			return
		}

		user, err := core.NewRecipient(username, pwhash, seedpwd, name)
		if err != nil {
			log.Println(err)
			erpc.ResponseHandler(w, erpc.StatusInternalServerError)
			return
		}

		erpc.MarshalSend(w, user)
	})
}

// validateRecipient validates a recipient on the platform
func validateRecipient() {
	http.HandleFunc(RecpRPC[3][0], func(w http.ResponseWriter, r *http.Request) {
		err := erpc.CheckGet(w, r)
		if err != nil {
			log.Println(err)
			return
		}

		prepRecipient, err := recpValidateHelper(w, r, RecpRPC[3][1:])
		if err != nil {
			return
		}
		erpc.MarshalSend(w, prepRecipient)
	})
}

// payback pays back towards an  invested order
func payback() {
	http.HandleFunc(RecpRPC[4][0], func(w http.ResponseWriter, r *http.Request) {
		err := erpc.CheckPost(w, r)
		if err != nil {
			log.Println(err)
			return
		}

		prepRecipient, err := recpValidateHelper(w, r, RecpRPC[4][1:])
		if err != nil {
			return
		}

		projIndexx := r.FormValue("projIndex")
		assetName := r.FormValue("assetName")
		seedpwd := r.FormValue("seedpwd")
		amountx := r.FormValue("amount")

		recpIndex := prepRecipient.U.Index
		projIndex, err := utils.ToInt(projIndexx)
		if err != nil {
			erpc.ResponseHandler(w, erpc.StatusBadRequest)
			return
		}
		amount, err := utils.ToFloat(amountx)
		if err != nil {
			erpc.ResponseHandler(w, erpc.StatusBadRequest)
			return
		}

		recipientSeed, err := wallet.DecryptSeed(prepRecipient.U.StellarWallet.EncryptedSeed, seedpwd)
		if err != nil {
			log.Println("did not decrypt seed", err)
			erpc.ResponseHandler(w, erpc.StatusBadRequest)
			return
		}

		err = core.Payback(recpIndex, projIndex, assetName, amount, recipientSeed)
		if err != nil {
			log.Println("did not payback", err)
			erpc.ResponseHandler(w, erpc.StatusInternalServerError)
			return
		}
		erpc.ResponseHandler(w, erpc.StatusOK)
	})
}

// chooseBlindAuction chooses a blind auction method to choose for the winner. Also commonly
// known as a 1st price auction.
func chooseBlindAuction() {
	http.HandleFunc(RecpRPC[8][0], func(w http.ResponseWriter, r *http.Request) {
		err := erpc.CheckGet(w, r)
		if err != nil {
			log.Println(err)
			return
		}

		recipient, err := recpValidateHelper(w, r, RecpRPC[8][1:])
		if err != nil {
			return
		}

		allContracts, err := core.RetrieveRecipientProjects(core.Stage2.Number, recipient.U.Index)
		if err != nil {
			log.Println("did not validate recipient projects", err)
			erpc.ResponseHandler(w, erpc.StatusInternalServerError)
			return
		}

		bestContract, err := core.SelectContractBlind(allContracts)
		if err != nil {
			log.Println("did not select contract", err)
			erpc.ResponseHandler(w, erpc.StatusInternalServerError)
			return
		}

		err = bestContract.SetStage(4)
		if err != nil {
			log.Println("did not set final project", err)
			erpc.ResponseHandler(w, erpc.StatusInternalServerError)
			return
		}

		erpc.ResponseHandler(w, erpc.StatusOK)
	})
}

// chooseVickreyAuction chooses a vickrey auction method to choose the winning contractor.
// also known as a second price auction
func chooseVickreyAuction() {
	http.HandleFunc(RecpRPC[9][0], func(w http.ResponseWriter, r *http.Request) {
		err := erpc.CheckGet(w, r)
		if err != nil {
			log.Println(err)
			return
		}

		recipient, err := recpValidateHelper(w, r, RecpRPC[9][1:])
		if err != nil {
			return
		}

		allContracts, err := core.RetrieveRecipientProjects(core.Stage2.Number, recipient.U.Index)
		if err != nil {
			log.Println("did not retrieve recipient projects", err)
			erpc.ResponseHandler(w, erpc.StatusInternalServerError)
			return
		}

		bestContract, err := core.SelectContractVickrey(allContracts)
		if err != nil {
			log.Println("did not select contract", err)
			erpc.ResponseHandler(w, erpc.StatusInternalServerError)
			return
		}

		err = bestContract.SetStage(4)
		if err != nil {
			log.Println("did not set final project", err)
			erpc.ResponseHandler(w, erpc.StatusInternalServerError)
			return
		}

		erpc.ResponseHandler(w, erpc.StatusOK)
	})
}

// chooseTimeAuction chooses the winning contractor based on least completion time
func chooseTimeAuction() {
	http.HandleFunc(RecpRPC[10][0], func(w http.ResponseWriter, r *http.Request) {
		err := erpc.CheckGet(w, r)
		if err != nil {
			log.Println(err)
			return
		}

		recipient, err := recpValidateHelper(w, r, RecpRPC[10][1:])
		if err != nil {
			return
		}

		allContracts, err := core.RetrieveRecipientProjects(core.Stage2.Number, recipient.U.Index)
		if err != nil {
			log.Println("did not retrieve recipient projects", err)
			erpc.ResponseHandler(w, erpc.StatusInternalServerError)
			return
		}

		bestContract, err := core.SelectContractTime(allContracts)
		if err != nil {
			log.Println("did not select contract", err)
			erpc.ResponseHandler(w, erpc.StatusInternalServerError)
			return
		}

		err = bestContract.SetStage(4)
		if err != nil {
			log.Println("did not set final project", err)
			erpc.ResponseHandler(w, erpc.StatusInternalServerError)
			return
		}

		erpc.ResponseHandler(w, erpc.StatusOK)
	})
}

// unlockOpenSolar unlocks a project which has just been invested in, signalling that the recipient
// has accepted the investment.
func unlockOpenSolar() {
	http.HandleFunc(RecpRPC[11][0], func(w http.ResponseWriter, r *http.Request) {
		err := erpc.CheckPost(w, r)
		if err != nil {
			log.Println(err)
			return
		}

		recipient, err := recpValidateHelper(w, r, RecpRPC[11][1:])
		if err != nil {
			return
		}

		seedpwd := r.FormValue("seedpwd")
		projIndexx := r.FormValue("projIndex")

		projIndex, err := utils.ToInt(projIndexx)
		if err != nil {
			log.Println("did not parse to integer", err)
			erpc.ResponseHandler(w, erpc.StatusBadRequest)
			return
		}

		err = core.UnlockProject(recipient.U.Username, recipient.U.Pwhash, projIndex, seedpwd)
		if err != nil {
			log.Println("did not unlock project", err)
			erpc.ResponseHandler(w, erpc.StatusInternalServerError)
			return
		}

		erpc.ResponseHandler(w, erpc.StatusOK)
	})
}

// addEmail adds an email address to the recipient's profile
func addEmail() {
	http.HandleFunc(RecpRPC[12][0], func(w http.ResponseWriter, r *http.Request) {
		err := erpc.CheckPost(w, r)
		if err != nil {
			log.Println(err)
			return
		}

		recipient, err := recpValidateHelper(w, r, RecpRPC[12][1:])
		if err != nil {
			return
		}

		email := r.FormValue("email")

		err = recipient.U.AddEmail(email)
		if err != nil {
			log.Println("did not add email", err)
			erpc.ResponseHandler(w, erpc.StatusBadRequest)
			return
		}
		erpc.ResponseHandler(w, erpc.StatusOK)
	})
}

// finalizeProject finalizes (ie moves from stage 2 to 3) a specific project
func finalizeProject() {
	http.HandleFunc(RecpRPC[13][0], func(w http.ResponseWriter, r *http.Request) {
		err := erpc.CheckPost(w, r)
		if err != nil {
			log.Println(err)
			return
		}

		_, err = recpValidateHelper(w, r, RecpRPC[13][1:])
		if err != nil {
			return
		}

		projIndexx := r.FormValue("projIndex")

		projIndex, err := utils.ToInt(projIndexx)
		if err != nil {
			log.Println("did not parse to integer", err)
			erpc.ResponseHandler(w, erpc.StatusBadRequest)
			return
		}

		project, err := core.RetrieveProject(projIndex)
		if err != nil {
			log.Println("did not retrieve project", err)
			erpc.ResponseHandler(w, erpc.StatusBadRequest)
			return
		}

		err = project.SetStage(4)
		if err != nil {
			log.Println("did not set final project", err)
			erpc.ResponseHandler(w, erpc.StatusInternalServerError)
			return
		}

		erpc.ResponseHandler(w, erpc.StatusOK)
	})
}

// originateProject originates (ie moves from stage 0 to 1) a project
func originateProject() {
	http.HandleFunc(RecpRPC[14][0], func(w http.ResponseWriter, r *http.Request) {
		err := erpc.CheckPost(w, r)
		if err != nil {
			log.Println(err)
			return
		}

		recipient, err := recpValidateHelper(w, r, RecpRPC[14][1:])
		if err != nil {
			return
		}

		projIndexx := r.FormValue("projIndex")

		projIndex, err := utils.ToInt(projIndexx)
		if err != nil {
			log.Println("did not parse to integer", err)
			erpc.ResponseHandler(w, erpc.StatusBadRequest)
			return
		}

		err = core.RecipientAuthorize(projIndex, recipient.U.Index)
		if err != nil {
			log.Println("did not authorize project", err)
			erpc.ResponseHandler(w, erpc.StatusInternalServerError)
			return
		}

		erpc.ResponseHandler(w, erpc.StatusOK)
	})
}

// calculateTrustLimit calculates the trust limit associated with a specific asset.
func calculateTrustLimit() {
	http.HandleFunc(RecpRPC[15][0], func(w http.ResponseWriter, r *http.Request) {
		err := erpc.CheckGet(w, r)
		if err != nil {
			log.Println(err)
			return
		}

		recipient, err := recpValidateHelper(w, r, RecpRPC[15][1:])
		if err != nil {
			return
		}

		assetName := r.URL.Query()["assetName"][0]
		trustLimit, err := xlm.GetAssetTrustLimit(recipient.U.StellarWallet.PublicKey, assetName)
		if err != nil {
			log.Println("did not get trust limit", err)
			erpc.ResponseHandler(w, erpc.StatusInternalServerError)
			return
		}

		erpc.MarshalSend(w, trustLimit)
	})
}

func setOneTimeUnlock() {
	http.HandleFunc(RecpRPC[17][0], func(w http.ResponseWriter, r *http.Request) {
		err := erpc.CheckPost(w, r)
		if err != nil {
			log.Println(err)
			return
		}

		// first validate the recipient or anyone would be able to set device ids
		prepRecipient, err := recpValidateHelper(w, r, RecpRPC[17][1:])
		if err != nil {
			return
		}

		projIndexx := r.FormValue("projIndex")
		seedpwd := r.FormValue("seedpwd")

		projIndex, err := utils.ToInt(projIndexx)
		if err != nil {
			erpc.ResponseHandler(w, erpc.StatusBadRequest)
			return
		}

		err = prepRecipient.SetOneTimeUnlock(projIndex, seedpwd)
		if err != nil {
			log.Println("did not set one time unlock", err)
			erpc.ResponseHandler(w, erpc.StatusInternalServerError)
			return
		}

		erpc.ResponseHandler(w, erpc.StatusOK)
	})
}
