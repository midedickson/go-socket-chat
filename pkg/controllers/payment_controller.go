package controllers

import "net/http"

func VerifyPayment(w http.ResponseWriter, r *http.Request) {
	// todo: implement this, it will be an endpoint the frontend will call by receiving the transaction reference
	// todo: find the user of the transaction, set ispaid to true, then call the MatchUser function
}
