package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/Double-DOS/go-socket-chat/db"
	"github.com/Double-DOS/go-socket-chat/pkg/match"
	"github.com/Double-DOS/go-socket-chat/pkg/websocket"
	"github.com/jmoiron/sqlx"
)

type PaystackVerifyResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Status    string `json:"status"`
		Reference string `json:"reference"`
		Amount    int    `json:"amount"`
		Customer  struct {
			Email string `json:"email"`
		} `json:"customer"`
	} `json:"data"`
}

func getPaystackSecretKey() string {
	key := os.Getenv("PAYSTACK_SECRET_KEY")
	if key == "" {
		panic("PAYSTACK_SECRET_KEY not set in environment")
	}
	return key
}

func updatePaymentStatus(db *sqlx.DB, reference string, status string) error {
	_, err := db.Exec(`
		UPDATE Payments
		SET status = $1
		WHERE transaction_reference = $2`,
		status, reference,
	)
	if err != nil {
		return fmt.Errorf("failed to update payment status: %v", err)
	}
	return nil
}
func findUserByReference(db *sqlx.DB, reference string) (*match.UserInfo, error) {
	var user match.UserInfo
	query := `
		SELECT u.*
		FROM Users u
		JOIN Payments p ON u.id = p.user_id
		WHERE p.transaction_reference = $1
	`
	err := db.Get(&user, query, reference)
	fmt.Println(err)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("transaction not found")
		}
		return nil, fmt.Errorf("failed to fetch user details: %v", err)
	}
	return &user, nil
}


func updateUser(db *sqlx.DB, userID int) (*match.UserInfo, error) {
	_, err := db.Exec(`
		UPDATE Users
		SET isPaid = true
		WHERE id = $1`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %v", err)
	}

	var user match.UserInfo
	err = db.Get(&user, `SELECT * FROM Users WHERE id = $1`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated user: %v", err)
	}

	return &user, nil
}

func VerifyPayment(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	reference := query.Get("reference")
	if reference == "" {
		http.Error(w, "Transaction reference is required", http.StatusBadRequest)
		return
	}

	db := db.DB
	if db == nil {
		http.Error(w, "Database connection failed", http.StatusInternalServerError)
		return
	}

	url := fmt.Sprintf("https://api.paystack.co/transaction/verify/%s", reference)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Authorization", "Bearer "+getPaystackSecretKey())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to send request to Paystack", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var paystackResponse PaystackVerifyResponse
	err = json.NewDecoder(resp.Body).Decode(&paystackResponse)
	if err != nil {
		http.Error(w, "Failed to parse response", http.StatusInternalServerError)
		return
	}

	if !paystackResponse.Status {
		http.Error(w, "Failed to verify transaction: "+paystackResponse.Message, http.StatusBadRequest)
		return
	}

	// Check if transaction was successful
	if paystackResponse.Data.Status != "success" {
		http.Error(w, "Payment not successful", http.StatusBadRequest)
		return
	}

	// Update payment status in the database
	err = updatePaymentStatus(db, reference, "PAID")
	if err != nil {
		http.Error(w, "Failed to update payment status", http.StatusInternalServerError)
		return
	}

	user, err := findUserByReference(db, reference)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	updatedUser, _ := updateUser(db, user.ID)
	// Call MatchUser function
	match.MatchUser(user)

	msg, _ := json.Marshal(websocket.ApiResponse{Success: true, Message: "Payment Verified Successfully!", Data: updatedUser})
	w.WriteHeader(http.StatusOK)
	w.Write(msg)	
}
