package controllers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/Double-DOS/go-socket-chat/pkg/match"
	"github.com/Double-DOS/go-socket-chat/pkg/websocket"
)

func GetRandomAnonNames(w http.ResponseWriter, r *http.Request) {
	// Allow requests from any origin

	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Allow specified HTTP methods

	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

	// Allow specified headers

	w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")

	if r.Method == http.MethodPost {
		defer r.Body.Close()
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			msg, _ := json.Marshal(websocket.ApiResponse{Success: true, Message: "error reading json body: " + err.Error()})
			w.Write(msg)
			return
		}
		var newUserInfo match.UserInfoDto
		err = json.Unmarshal(bodyBytes, &newUserInfo)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			msg, _ := json.Marshal(websocket.ApiResponse{Success: true, Message: "error reading json body: " + err.Error()})
			w.Write(msg)
			return
		}
		newUser := newUserInfo.NewUserInfo()
		msg, _ := json.Marshal(websocket.ApiResponse{Success: true, Message: "New user registered Successfully!", Data: newUser})
		w.WriteHeader(http.StatusOK)
		w.Write(msg)

	}
}