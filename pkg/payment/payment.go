package payment

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type PaystackInitializeRequest struct {
	Email    string                 `json:"email"`
	Amount   int                    `json:"amount"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type PaystackInitializeResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		AuthorizationUrl string `json:"authorization_url"`
		AccessCode       string `json:"access_code"`
		Reference        string `json:"reference"`
	} `json:"data"`
}

var httpClient = &http.Client{}

const paystackSecretKeyEnv = "PAYSTACK_SECRET_KEY"

func getPaystackSecretKey() string {
	key := os.Getenv(paystackSecretKeyEnv)
	if key == "" {
		panic("PAYSTACK_SECRET_KEY not set in environment")
	}
	return key
}

func savePayment(db *sql.DB, userID int, amount int, transactionReference string) error {
	_, err := db.Exec(`
		INSERT INTO Payments (user_id, amount, status, transaction_reference)
		VALUES ($1, $2, $3, $4)`,
		userID, amount, "PENDING", transactionReference,
	)
	if err != nil {
		return fmt.Errorf("failed to save payment record: %v", err)
	}
	return nil
}

func InitializePayment(db *sql.DB, userID int, email string, amount int, metadata map[string]interface{}) (string, error) {
	if email == "" {
		return "", fmt.Errorf("email is required")
	}
	if amount <= 0 {
		return "", fmt.Errorf("amount must be greater than 0")
	}

	requestBody := PaystackInitializeRequest{
		Email:    email,
		Amount:   amount,
		Metadata: metadata,
	}

	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	url := "https://api.paystack.co/transaction/initialize"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestJSON))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+getPaystackSecretKey())
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request to Paystack: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	var paystackResponse PaystackInitializeResponse
	err = json.Unmarshal(body, &paystackResponse)
	if err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	if !paystackResponse.Status {
		return "", fmt.Errorf("paystack error: %s", paystackResponse.Message)
	}

	transactionReference := paystackResponse.Data.Reference
	err = savePayment(db, userID, amount, transactionReference)
	if err != nil {
		return "", err
	}

	return paystackResponse.Data.AuthorizationUrl, nil
}
