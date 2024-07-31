package main

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/blue-davinci/aggregate/internal/data"
	"github.com/blue-davinci/aggregate/internal/validator"
	"github.com/hashicorp/go-retryablehttp"
)

func (app *application) initializeTransactionHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		PlanID      int64  `json:"plan_id"`
		Amount      int64  `json:"amount"`
		CallBackURL string `json:"callback_url"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	user := app.contextGetUser(r)
	transactionData := &data.TransactionData{
		User_ID:     user.ID,
		Plan_ID:     input.PlanID,
		Amount:      input.Amount,
		Email:       user.Email,
		CallBackURL: input.CallBackURL,
	}
	v := validator.New()
	if data.ValidateTransactionData(v, transactionData); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	if transactionData.CallBackURL == "" {
		transactionData.CallBackURL = app.config.frontend.callback_url
	}
	initializeResponse, err := app.transactionClient(transactionData, app.config.paystack.initializationurl, data.PaymentOperationInitialize)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusCreated, envelope{"initialization": initializeResponse.InitializeResponse, "transaction_data": transactionData}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) verifyTransactionHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Reference string `json:"reference"`
		Plan_ID   int64  `json:"plan_id"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	user := app.contextGetUser(r)
	transactionData := &data.TransactionData{
		User_ID:   user.ID,
		Plan_ID:   input.Plan_ID,
		Email:     user.Email,
		Reference: input.Reference,
	}
	v := validator.New()
	if data.ValidateVerificationData(v, transactionData); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	verificationUrlEndpoint := fmt.Sprintf("%s%s", app.config.paystack.verificationurl, transactionData.Reference)
	app.logger.PrintInfo("verification url", map[string]string{"url": verificationUrlEndpoint})
	verifyResponse, err := app.transactionClient(transactionData, verificationUrlEndpoint, data.PaymentOperationVerify)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	app.logger.PrintInfo("verify response", map[string]string{
		"status":    fmt.Sprintf("%t", verifyResponse.VerifyResponse.Status),
		"message":   verifyResponse.VerifyResponse.Message,
		"card_type": verifyResponse.VerifyResponse.Data.Authorization.CardType,
	})
	err = app.writeJSON(w, http.StatusOK, envelope{"verification": verifyResponse, "transaction_data": transactionData}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) transactionClient(transactionData *data.TransactionData, payment_url string, paymentOperation data.PaymentOperation) (*data.UnifiedPaymentResponse, error) {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 5
	var body *bytes.Buffer
	var req *retryablehttp.Request
	var err error

	if paymentOperation == data.PaymentOperationInitialize {
		jsonData, err := app.covertToByteArray(transactionData)
		if err != nil {
			app.logger.PrintError(err, nil)
			return nil, err
		}
		body = bytes.NewBuffer(jsonData)
		req, err = retryablehttp.NewRequest("POST", payment_url, body)
		if err != nil {
			app.logger.PrintError(err, nil)
			return nil, err
		}
	} else {
		req, err = retryablehttp.NewRequest("GET", payment_url, nil)
		if err != nil {
			app.logger.PrintError(err, nil)
			return nil, err
		}
	}

	app.logger.PrintInfo("secret key: ", map[string]string{"key": app.config.paystack.secretkey})
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", app.config.paystack.secretkey))
	req.Header.Set("Content-Type", "application/json")
	resp, err := retryClient.Do(req)
	if err != nil {
		app.logger.PrintError(err, nil)
		return nil, err
	}
	defer resp.Body.Close()

	var paymentRequest any
	if paymentOperation == data.PaymentOperationInitialize {
		paymentRequest = &data.InitializeResponse{}
	} else {
		paymentRequest = &data.VerifyResponse{}
	}
	app.logger.PrintInfo("payment request", map[string]string{"request": fmt.Sprintf("%+v", paymentRequest)})
	err = app.readJSONFromReader(resp.Body, &paymentRequest)
	if err != nil {
		return nil, err
	}

	if paymentOperation == data.PaymentOperationInitialize {
		return &data.UnifiedPaymentResponse{InitializeResponse: *paymentRequest.(*data.InitializeResponse)}, nil
	}
	return &data.UnifiedPaymentResponse{VerifyResponse: *paymentRequest.(*data.VerifyResponse)}, nil
}
