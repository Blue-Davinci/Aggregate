package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/blue-davinci/aggregate/internal/data"
	"github.com/blue-davinci/aggregate/internal/validator"
	"github.com/hashicorp/go-retryablehttp"
)

// getPaymentPlansHandler() is a handler that gets all the available subscription plans.
// it's a simple route, all we do is get the plans from the database and write them to the client.
func (app *application) getPaymentPlansHandler(w http.ResponseWriter, r *http.Request) {
	plans, err := app.models.Payments.GetPaymentPlans()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"plans": plans}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// initializeTransactionHandler() is a handler that creates an intent for the transaction
// we get in the plan ID, amount in cts and a callback URL. If the callback URL is not
// provided, we default to the internal callback URL. The plan ID keeps track of the plan
// we will be using incase we will ever want to save an intent, which we don't now.
// We validate the transaction data and then send the data to the transaction client
// which should send a post request and get back a response which contains a reference as well
// and more importantly the authorization URL. We then write the response to the client.
func (app *application) initializeTransactionHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		PlanID      int32  `json:"plan_id"`
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
	// we verify that the plan is a valid one
	plan, err := app.models.Payments.GetPaymentPlanByID(transactionData.Plan_ID)
	if err != nil {
		switch {
		// if the plan is not found, we return a 404 error
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// we now set the amount into our transaction data
	transactionData.Amount = plan.Price * 100 // we multiply it by 100 since the default for paystack is in cents
	initializeResponse, err := app.transactionClient(transactionData, app.config.paystack.initializationurl, data.PaymentOperationInitialize)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// We send back the transaction Data incase the frontend needs it as well as the initialize response which
	// the frontend will require, using both the auth URL and the reference.
	err = app.writeJSON(w, http.StatusCreated, envelope{"initialization": initializeResponse.InitializeResponse, "transaction_data": transactionData}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// verifyTransactionHandler() is a handler that verifies a transaction. We get the reference
// and the plan ID. We validate the transaction data and then send the data to the transaction
// client which should send a get request and get back a response which contains the status of the
// transaction, the message and the card type in additiona to a whole bevy of info. We will
// only save this data if it was actually successful or send back an error if not
func (app *application) verifyTransactionHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Reference string `json:"reference"`
		Plan_ID   int32  `json:"plan_id"`
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

	// we get the plan to fill in required data
	plan, err := app.models.Payments.GetPaymentPlanByID(transactionData.Plan_ID)
	if err != nil {
		switch {
		// if the plan is not found, we return a 404 error
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// we now send the transaction data to the transaction client to verify the transaction
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
	// we need to check verifyResponse.data.status to see if its "success" or "failed"
	// and verifyResponse.data.gateway_response to see if its  "Approved" or "Declined"
	if verifyResponse.VerifyResponse.Data.Status != "success" && verifyResponse.VerifyResponse.Data.GatewayResponse != "Approved" {
		// we assume this is a failed transaction and return a 400 error
		app.badRequestResponse(w, r, data.ErrTransactionDeclined)
		return
	}
	// if the transaction was successful, we save the transaction data to the database
	payment_detail := &data.Payment_Details{
		User_ID:            user.ID,
		Plan_ID:            transactionData.Plan_ID,
		Start_Date:         time.Now().UTC(),
		End_Date:           app.returnEndDate(plan.Duration, time.Now().UTC()),
		Price:              verifyResponse.VerifyResponse.Data.Amount / 100,
		Status:             "active",
		TransactionID:      verifyResponse.VerifyResponse.Data.ID,
		Payment_Method:     verifyResponse.VerifyResponse.Data.Channel,
		Authorization_Code: verifyResponse.VerifyResponse.Data.Authorization.AuthorizationCode,
		Card_Last4:         verifyResponse.VerifyResponse.Data.Authorization.Last4,
		Card_Exp_Month:     verifyResponse.VerifyResponse.Data.Authorization.ExpMonth,
		Card_Exp_Year:      verifyResponse.VerifyResponse.Data.Authorization.ExpYear,
		Card_Type:          verifyResponse.VerifyResponse.Data.Authorization.CardType,
		Currency:           verifyResponse.VerifyResponse.Data.Currency,
	}

	err = app.models.Payments.CreateSubscription(payment_detail)
	// if we get a constraint validation on the transaction ID, we return a 400 error
	// as we know we have already processed the same transaction.
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateTransaction):
			v.AddError("transaction", "a transaction with this reference already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// ToDo: Create go routine to send an email of payment success to the user, Maybe a reciept.
	app.background(func() {
		data := map[string]any{
			"UserName":        user.Name,
			"TransactionID":   payment_detail.TransactionID,
			"PlanName":        plan.Name,
			"AmountPaid":      payment_detail.Price,
			"Currency":        payment_detail.Currency,
			"PaymentMethod":   payment_detail.Payment_Method,
			"Date":            payment_detail.Start_Date,
			"TransactionDate": app.formatDate(verifyResponse.VerifyResponse.Data.TransactionDate),
			"GrandTotal":      payment_detail.Price,
		}
		app.logger.PrintInfo("Data DATA:", map[string]string{
			"Date":      verifyResponse.VerifyResponse.Data.TransactionDate,
			"Converted": app.formatDate(verifyResponse.VerifyResponse.Data.TransactionDate),
		})
		err = app.mailer.Send(user.Email, "subscription_reciept.tmpl", data)
		if err != nil {
			app.logger.PrintError(err, nil)
		}
	})
	// We send back the transaction Data incase the frontend needs it as well as the verify response which
	// the frontend may require.
	err = app.writeJSON(w, http.StatusOK, envelope{"payment_details": payment_detail, "transaction_data": transactionData}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// transactionClient() sets up the client side of this module. The client is different from the scraper
// and performs tasks differently thus we just made a none-unified one. We set up a retryable client
// which sets up the body based on the type of transaction request we are doing. If a transaction request is
// an initialization request i.e PaymentOperationInitialize, we use a POST request to send the transaction
// DATA to the payment URL while If it is a verification request i.e PaymentOperationVerify, we use a GET request
// to get the transaction data from the payment URL. We make sure to set relevant headers for each request,
// most important of which is the PAYSTACK secret KEY.
// the responses are written in a unified struct and returned to the callers which can access them via
// their appropriate fields.
func (app *application) transactionClient(transactionData *data.TransactionData, payment_url string, paymentOperation data.PaymentOperation) (*data.UnifiedPaymentResponse, error) {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 5
	var body *bytes.Buffer
	var req *retryablehttp.Request
	var err error
	// if the operation is an initialization, we do a quick byte conversion of the
	// body and set up the request to be a POST request
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

	// we set the headers for the request, using our token saved in our env file
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", app.config.paystack.secretkey))
	req.Header.Set("Content-Type", "application/json")
	resp, err := retryClient.Do(req)
	if err != nil {
		app.logger.PrintError(err, nil)
		return nil, err
	}
	defer resp.Body.Close()

	// prep the response reciever
	var paymentRequest any
	if paymentOperation == data.PaymentOperationInitialize {
		paymentRequest = &data.InitializeResponse{}
	} else {
		paymentRequest = &data.VerifyResponse{}
	}
	//app.logger.PrintInfo("payment request", map[string]string{"request": fmt.Sprintf("%+v", paymentRequest)})
	// get the response and decode it into our reciever
	err = app.readJSONFromReader(resp.Body, &paymentRequest)
	if err != nil {
		return nil, err
	}

	if paymentOperation == data.PaymentOperationInitialize {
		return &data.UnifiedPaymentResponse{InitializeResponse: *paymentRequest.(*data.InitializeResponse)}, nil
	}
	return &data.UnifiedPaymentResponse{VerifyResponse: *paymentRequest.(*data.VerifyResponse)}, nil
}
