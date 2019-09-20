package rpc

import (
	"log"
	"net/http"

	erpc "github.com/Varunram/essentials/rpc"
	core "github.com/test/blah/core"
)

func setupPublicRoutes() {
	getAllInvestorsPublic()      // GET
	getAllRecipientsPublic()     // GET
	getInvTopReputationPublic()  // GET
	getRecpTopReputationPublic() // GET
}

var PublicRpc = map[int][]string{
	1: []string{"/public/investor/all"},
	2: []string{"/public/recipient/all"},
	3: []string{"/public/recipient/reputation/top"},
	4: []string{"/public/investor/reputation/top"},
}

// SnInvestor defines a sanitized investor
type SnInvestor struct {
	Name             string
	InvestedProjects []string
	AmountInvested   float64
	InvestedBonds    []string
	InvestedCoops    []string
	PublicKey        string
	Reputation       float64
}

// SnRecipient defines a sanitized recipient
type SnRecipient struct {
	Name                  string
	PublicKey             string
	ReceivedSolarProjects []string
	Reputation            float64
}

// SnUser defines a sanitized user
type SnUser struct {
	Name       string
	PublicKey  string
	Reputation float64
}

// sanitizeInvestor removes sensitive fields from the investor struct
func sanitizeInvestor(investor core.Investor) SnInvestor {
	var sanitize SnInvestor
	sanitize.Name = investor.U.Name
	sanitize.InvestedProjects = investor.InvestedProjects
	sanitize.AmountInvested = investor.AmountInvested
	sanitize.PublicKey = investor.U.StellarWallet.PublicKey
	sanitize.Reputation = investor.U.Reputation
	return sanitize
}

// sanitizeRecipient removes sensitive fields from the recipient struct
func sanitizeRecipient(recipient core.Recipient) SnRecipient {
	var sanitize SnRecipient
	sanitize.Name = recipient.U.Name
	sanitize.PublicKey = recipient.U.StellarWallet.PublicKey
	sanitize.Reputation = recipient.U.Reputation
	sanitize.ReceivedSolarProjects = recipient.ReceivedSolarProjects
	return sanitize
}

// sanitizeAllInvestors sanitizes an array of investors
func sanitizeAllInvestors(investors []core.Investor) []SnInvestor {
	var arr []SnInvestor
	for _, elem := range investors {
		arr = append(arr, sanitizeInvestor(elem))
	}
	return arr
}

// sanitizeAllRecipients sanitizes an array of recipients
func sanitizeAllRecipients(recipients []core.Recipient) []SnRecipient {
	var arr []SnRecipient
	for _, elem := range recipients {
		arr = append(arr, sanitizeRecipient(elem))
	}
	return arr
}

// getAllInvestors gets a list of all the investors in the database
func getAllInvestorsPublic() {
	http.HandleFunc(PublicRpc[1][0], func(w http.ResponseWriter, r *http.Request) {
		err := erpc.CheckGet(w, r)
		if err != nil {
			log.Println(err)
			return
		}
		investors, err := core.RetrieveAllInvestors()
		if err != nil {
			log.Println("did not retrieve all investors", err)
			erpc.ResponseHandler(w, erpc.StatusInternalServerError)
			return
		}
		sInvestors := sanitizeAllInvestors(investors)
		erpc.MarshalSend(w, sInvestors)
	})
}

// getAllRecipients gets a list of all the investors in the database
func getAllRecipientsPublic() {
	http.HandleFunc(PublicRpc[2][0], func(w http.ResponseWriter, r *http.Request) {
		err := erpc.CheckGet(w, r)
		if err != nil {
			log.Println(err)
			return
		}
		recipients, err := core.RetrieveAllRecipients()
		if err != nil {
			log.Println("did not retrieve all recipients", err)
			erpc.ResponseHandler(w, erpc.StatusInternalServerError)
			return
		}
		sRecipients := sanitizeAllRecipients(recipients)
		erpc.MarshalSend(w, sRecipients)
	})
}

// getRecpTopReputationPublic gets a list of the recipients sorted by descending order of reputation
func getRecpTopReputationPublic() {
	http.HandleFunc(PublicRpc[3][0], func(w http.ResponseWriter, r *http.Request) {
		err := erpc.CheckGet(w, r)
		if err != nil {
			log.Println(err)
			return
		}
		allRecps, err := core.TopReputationRecipients()
		if err != nil {
			log.Println("did not retrieve all top reputaiton recipients", err)
			erpc.ResponseHandler(w, erpc.StatusInternalServerError)
			return
		}
		sRecipients := sanitizeAllRecipients(allRecps)
		erpc.MarshalSend(w, sRecipients)
	})
}

// getInvTopReputationPublic gets a list of the investors sorted by descending order of reputation
func getInvTopReputationPublic() {
	http.HandleFunc(PublicRpc[4][0], func(w http.ResponseWriter, r *http.Request) {
		err := erpc.CheckGet(w, r)
		if err != nil {
			log.Println(err)
			return
		}
		allInvs, err := core.TopReputationInvestors()
		if err != nil {
			log.Println("did not retrieve all top reputation investors", err)
			erpc.ResponseHandler(w, erpc.StatusInternalServerError)
			return
		}
		sInvestors := sanitizeAllInvestors(allInvs)
		erpc.MarshalSend(w, sInvestors)
	})
}
