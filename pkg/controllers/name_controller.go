package controllers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/Double-DOS/go-socket-chat/pkg/match"
	"github.com/Double-DOS/go-socket-chat/pkg/websocket"
	"github.com/Double-DOS/randommer-go"
)

func GetRandomAnonNames(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		names := randommer.GetRandomNames("firstname", 1)

		msg, err := json.Marshal(websocket.ApiResponse{Success: true, Message: "Fetched Random Name Successfully", Data: names[0]})
		if err == nil {
			w.WriteHeader(http.StatusOK)
			w.Write(msg)
		} else {
			log.Printf("err: %s", err)
			msg, _ = json.Marshal(websocket.ApiResponse{Success: false, Message: "Fetching Random Name Failed", Data: nil})
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(msg)
		}
	}
	if r.Method == http.MethodPost {
		defer r.Body.Close()
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			msg, _ := json.Marshal(websocket.ApiResponse{Success: true, Message: "error reading json body: " + err.Error()})
			w.Write(msg)
		}
		var newUserInfo match.UserInfoDto
		err = json.Unmarshal(bodyBytes, &newUserInfo)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			msg, _ := json.Marshal(websocket.ApiResponse{Success: true, Message: "error reading json body: " + err.Error()})
			w.Write(msg)
		}
		newUser := newUserInfo.NewUserInfo()
		msg, _ := json.Marshal(websocket.ApiResponse{Success: true, Message: "New user registered Successfully!", Data: newUser})
		w.WriteHeader(http.StatusOK)
		w.Write(msg)

	}
}
