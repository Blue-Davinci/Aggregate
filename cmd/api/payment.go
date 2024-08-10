package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/blue-davinci/aggregate/internal/data"
	"github.com/blue-davinci/aggregate/internal/validator"
	"github.com/google/uuid"
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

// getPendingChallengedTransactionsByUser() is a handler that gets all the pending challenged transactions
// by a specific user. This endpoint is used to allow the frontend to poll for any challenged transactions
// that are pending. We get the user from the context and then get the transactions from the database.
func (app *application) getPendingChallengedTransactionsByUser(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)
	challengedTransactions, err := app.models.Payments.GetPendingChallengedTransactionsByUser(user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"challenged_transactions": challengedTransactions}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// GetAllSubscriptionsByIDHandler() returns the payment/subscription history of a user.
// Returning all transactions made by the user in addition to each subscription's info.
func (app *application) getAllSubscriptionsByIDHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		data.Filters
	}
	//validate if queries are provided
	v := validator.New()
	// Call r.URL.Query() to get the url.Values map containing the query string data.
	qs := r.URL.Query()
	//get the pagesizes as ints and set to the embedded struct
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	// get the sort values falling back to "id" if it is not provided
	input.Filters.Sort = app.readString(qs, "sort", "id")
	// Add the supported sort values for this endpoint to the sort safelist.
	input.Filters.SortSafelist = []string{"id", "-id"}
	// Perform validation
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	subscriptions, metadata, err := app.models.Payments.GetAllSubscriptionsByID(app.contextGetUser(r).ID, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"subscriptions": subscriptions, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// updateSubscriptionStatusForUserHandler() is a handler that updates the status of a subscription
// moreso designed to register cancelled subscriptions from a users account. We get the subscription ID
func (app *application) updateSubscriptionStatusForUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		SubscriptionID uuid.UUID `json:"subscription_id"`
		Status         string    `json:"status"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	// make a subscription
	subscription := &data.Subscription{
		ID:     input.SubscriptionID,
		Status: input.Status,
	}
	// validate the Subscription ID and the status
	v := validator.New()
	if data.ValidateSubscriptionStatus(v, subscription); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// we are okay so let us feed the ID and status
	err = app.models.Payments.UpdateSubscriptionStatus(subscription.ID, subscription.Status, app.contextGetUser(r).ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"subscription": subscription}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// updateChallengedTransactionStatus() updates the status of a challenged transaction
// designed to set to failed or successful based on the user's resolve
func (app *application) updateChallengedTransactionStatus(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ChallengedTransactionID     int64  `json:"challenged_transaction_id"`
		ChallengedTransactionStatus string `json:"challenged_transaction_status"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	challengedTransaction := &data.ChallengedTransaction{
		ID:     input.ChallengedTransactionID,
		Status: input.ChallengedTransactionStatus,
	}
	// validate the Comment ID and Comment text
	v := validator.New()
	if data.ValidateChallengedTransaction(v, challengedTransaction); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// we are okay so let us feed the ID, status and the user behind the request
	updatedChallangedTransaction, err := app.models.Payments.UpdateChallengedTransactionStatus(challengedTransaction.ID, app.contextGetUser(r).ID, challengedTransaction.Status)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"challenged_transaction": updatedChallangedTransaction}, nil)
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
	app.logger.PrintInfo("callback url", map[string]string{"url": transactionData.CallBackURL})
	// check if a user has an existing subscription, if they do, we return a 409 conflict error
	subscription, err := app.models.Payments.GetSubscriptionByID(user.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrSubscriptionNotFound):
			// we ignore this error, means a user deos not have a subscription
		default:
			app.serverErrorResponse(w, r, err)
			return
		}
	}
	// if the user has a subscription, we return a 409 conflict error
	if subscription != nil {
		v.AddError("transaction", "user already has a subscription")
		app.failedConstraintValidation(w, r, v.Errors)
		return
	}
	// if user has no sub, we verify that the provided plan is a valid one
	plan, err := app.models.Payments.GetPaymentPlanByID(transactionData.Plan_ID)
	if err != nil {
		switch {
		// if the plan is not found, we return a 404 error
		case errors.Is(err, data.ErrPaymentPlanNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	if plan.Price != transactionData.Amount {
		// if the amount is not the same as the plan price, we return a 400 error
		app.badRequestResponse(w, r, errors.New("we could not process the data due to a discrepancy"))
		return
	}
	// we now set the amount into our transaction data
	transactionData.Amount *= 100 * 100 // we multiply it by 100*100 since the default for paystack is in cents
	app.logger.PrintInfo("amount", map[string]string{"amount": fmt.Sprintf("%d", transactionData.Amount),
		"plan":       plan.Name,
		"plan price": fmt.Sprintf("%d", plan.Price)})
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
		case errors.Is(err, data.ErrPaymentPlanNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	//quick fill for the transactio data
	transactionData.Amount = plan.Price
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
		failedTransaction := fmt.Sprintf("error: %s\nplan: %s\nemail: %s", data.ErrTransactionDeclined.Error(), plan.Name, user.Email)
		app.badRequestResponse(w, r, errors.New(failedTransaction))
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

	err = app.createSubscriptionHandler(payment_detail, plan.Name, user.Name, user.Email, verifyResponse.VerifyResponse.Data.TransactionDate)
	// if we get a constraint validation on the transaction ID, we return a 400 error
	// as we know we have already processed the same transaction.
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateTransaction):
			v.AddError("transaction", "[RP-T] An error occurred with your transaction. Please try again. If the issue persists, please contact support with this message.")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// We send back the transaction and Payment details Data back incase the frontend needs it
	// maybe for items such as reciept generation etc.
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
	if paymentOperation == data.PaymentOperationInitialize || paymentOperation == data.PaymentOperationRecurring {
		app.logger.PrintInfo("transaction data", map[string]string{"Transaction data Amount": fmt.Sprintf("%d", transactionData.Amount)})
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

// startAutoSubscriptionHandler() is a handler that starts the auto subscription handler.
// It sets up a cron job that runs every x minutes to check for subscriptions that are due for renewal.
// The default time will run every half a day i.e every 720 minutes.
func (app *application) startPaymentSubscriptionHandler() {
	app.logger.PrintInfo("Starting autosubscription jobs...", map[string]string{
		"Interval": fmt.Sprintf("%d", app.config.paystack.autosubscriptioninterval),
	})
	// get our cron job interval autosubscriptioninterval from the flag
	//app.config.paystack.autosubscriptioninterval
	autoSubscriptionInterval := fmt.Sprintf("*/%d * * * *", 1)
	// get our expiry check interval from the flag
	//app.config.paystack.checkexpiredsubscriptioninterval
	expiryCheckInterval := fmt.Sprintf("*/%d * * * *", 1)

	// add the autosubscription handler to the cron job
	_, err := app.config.paystack.cronJob.AddFunc(autoSubscriptionInterval, app.autoSubscriptionHandler)
	if err != nil {
		app.logger.PrintError(err, map[string]string{
			"Error": "Error adding autosubscription job",
		})
	}

	// add the subscription expiry checker to the cron job
	_, err = app.config.paystack.cronJob.AddFunc(expiryCheckInterval, app.updateExpiredSubscriptionHandler)
	if err != nil {
		app.logger.PrintError(err, map[string]string{
			"Error": "Error adding expiry check job",
		})
	}

	// since the cron jobs don't start immediately, we need to start some of our handlers
	// manually as sort of a first run
	// ------------------------------------------
	// perform an expired check before we start the cron
	app.updateExpiredSubscriptionHandler()
	// perform and check for any expired subscriptions to begin autosubscription
	app.autoSubscriptionHandler()

	// start the cron scheduler
	app.config.paystack.cronJob.Start()
}

// autoSubscriptionHandler() should handle recurring charges on a subscribers account.
// It should start by selecting all accounts whose subscription is due for renewal.
// For each of those accounts, we will need the account email, subscription auth
// and the amount to recharge. We will then use this information to send the data
// using our pre-existing client to the initialization endpoint.
// We will then check the status, checking if the transaction is paused. If it's paused
// or even failed, we will add it to our table from where the frontend will poll
// and send a notification to the user. If the transaction is successful, we will
// add a new subscription and update the challanged item, whether succesful or failed.
// If successful, send an acknowledgement email to user, if not, send a checkup email instead.
// No Errors should really be treated as panics, so this methods weill just log the error and proceed
func (app *application) autoSubscriptionHandler() {
	app.logger.PrintInfo("auto subscription handler", nil)
	//setup filters
	var input struct{ data.Filters }
	input.Filters.Page = app.readInt(nil, "page", 1, nil)
	input.Filters.PageSize = app.readInt(nil, "page_size", 5, nil)
	// get all subscriptions that are due for renewal
	subscriptions, _, err := app.models.Payments.GetAllExpiredSubscriptions(input.Filters)
	if err != nil {
		app.logger.PrintError(err, nil)
	}
	// for each subscription, we will send a request to the payment client,
	// using the charge authorization endpoint instead.
	for _, subscription := range subscriptions {
		app.logger.PrintInfo("subscription", map[string]string{"email": subscription.User_Email, "plan": subscription.Authorization_Code})
		// each item is a subscription, so we concert to a transacyion data
		err := app.processRecurringSubscription(subscription)
		if err != nil {
			app.logger.PrintError(err, nil)
		}
	}
}

// processRecurringSubscription() is the main handler that processes recurring subscriptions.
// That is, it checks for any and all subscriptions that are expired, and for each of them
// It proceeds to perform a subsequent charge of them provided they are not cancelled.
func (app *application) processRecurringSubscription(subscription *data.RecurringSubscription) error {
	app.logger.PrintInfo("Processing recurring subscription", map[string]string{"email": subscription.User_Email, "plan": subscription.Authorization_Code})
	transactionData := &data.TransactionData{
		Email:              subscription.User_Email,
		Amount:             subscription.Subscription.Price * 100,
		Authorization_Code: subscription.Authorization_Code,
		CallBackURL:        fmt.Sprintf("%s%s%d", app.config.frontend.callback_url, "/health?plan_id=", subscription.Subscription.Plan_ID),
	}
	app.logger.PrintInfo(">>>>> Price", map[string]string{"Price": fmt.Sprintf("%d", transactionData.Amount)})
	// We need  to FIRST check if a user has a challanged transaction, if they do,
	// there is no need to process it again and we return and proceed with the next subscription.
	challengedTransaction, err := app.models.Payments.GetPendingChallengedTransactionBySubscriptionID(subscription.Subscription.ID)
	if err != nil {
		switch {
		// check if the error is a not found error, if it is, we ignore it
		case errors.Is(err, data.ErrChallangedTransactionNotFound):
			// we ignore this error, means a user does not have a challenged transaction
		default:
			return err
		}
	}
	// if there is a pending challanged transaction, we return nil and proceed to the next subscription
	if challengedTransaction != nil {
		return nil
	}

	// it will use a post request, and will feed a new charge auth url instead
	chargeAuthResponse, err := app.transactionClient(transactionData, app.config.paystack.chargeauthorizationurl, data.PaymentOperationRecurring)
	if err != nil {
		return err
	}
	// if the transaction was challanged, it will have 2 distinct items, Paused=true + authorization_url
	// when it's paused, we will need to add this transaction to the challanged transaction table
	if chargeAuthResponse.VerifyResponse.Data.Paused {
		err := app.createChallengedTransaction(subscription, &chargeAuthResponse.VerifyResponse)
		if err != nil {
			return err
		}
	}
	// Get our plan
	plan, err := app.models.Payments.GetPaymentPlanByID(subscription.Subscription.Plan_ID)
	if err != nil {
		return err
	}
	// Create a payment Detail as we will use this whether it's successful or not
	paymentDetails := &data.Payment_Details{
		ID:                 subscription.Subscription.ID,
		User_ID:            subscription.User_ID,
		Plan_ID:            subscription.Subscription.Plan_ID,
		Start_Date:         time.Now().UTC(),
		End_Date:           app.returnEndDate(plan.Duration, time.Now().UTC()),
		Price:              subscription.Subscription.Price,
		Authorization_Code: subscription.Authorization_Code,
		TransactionID:      chargeAuthResponse.VerifyResponse.Data.ID,
		Payment_Method:     chargeAuthResponse.VerifyResponse.Data.Channel,
		Card_Last4:         chargeAuthResponse.VerifyResponse.Data.Authorization.Last4,
		Card_Exp_Month:     chargeAuthResponse.VerifyResponse.Data.Authorization.ExpMonth,
		Card_Exp_Year:      chargeAuthResponse.VerifyResponse.Data.Authorization.ExpYear,
		Card_Type:          chargeAuthResponse.VerifyResponse.Data.Authorization.CardType,
		Currency:           chargeAuthResponse.VerifyResponse.Data.Currency,
	}
	// if the transaction was not successful, we will add it to the failed transaction table
	if chargeAuthResponse.VerifyResponse.Data.Status != "success" && chargeAuthResponse.VerifyResponse.Data.GatewayResponse != "Approved" {
		err := app.createFailedTransactionHandler(paymentDetails, *chargeAuthResponse.VerifyResponse.Data.Message,
			chargeAuthResponse.VerifyResponse.Data.Reference,
			chargeAuthResponse.VerifyResponse.Data.GatewayResponse)
		app.logger.PrintInfo("Adding failed transaction", map[string]string{"Status": paymentDetails.Status, "id": fmt.Sprintf("%d", paymentDetails.ID)})
		if err != nil {
			return err
		}
	}
	// if the transaction was successful, we save the transaction data to the database
	err = app.createSubscriptionHandler(paymentDetails, plan.Name, subscription.User_Name,
		subscription.User_Email, chargeAuthResponse.VerifyResponse.Data.TransactionDate)
	if err != nil {
		return err
	}
	// log this successful transaction
	app.logger.PrintInfo("Adding Successful subscription", map[string]string{"Payment Method": paymentDetails.Payment_Method,
		"Transaction id": fmt.Sprintf("%d", paymentDetails.TransactionID)})

	// now perform an update on the old subscription setting it to renewed for historical reasons
	// as we will now have a new subscription for this specific user which we added above.
	err = app.models.Payments.UpdateSubscriptionStatus(subscription.Subscription.ID, data.PaymentStatusRenewed, subscription.User_ID)
	if err != nil {
		return err
	}
	// log a successful update
	app.logger.PrintInfo("Successfully Updated a subscription",
		map[string]string{"Subscription ID": subscription.Subscription.ID.String()})

	// we are good so return back a nil error
	return nil
}

// createSubscriptionHandler() Creates a subscription taking in payment details and user information
// This handler sends a succesfull transaction email to the user as well if the data is saved succesfully.
func (app *application) createSubscriptionHandler(payment_detail *data.Payment_Details, plan_name, user_name, user_email, transactionDate string) error {
	err := app.models.Payments.CreateSubscription(payment_detail)
	// if we get a constraint validation on the transaction ID, we return a 400 error
	// as we know we have already processed the same transaction.
	if err != nil {
		return err
	}
	// We are good, so we send an email acknowledgment to the user.
	app.background(func() {
		data := map[string]any{
			"UserName":        user_name,
			"TransactionID":   payment_detail.TransactionID,
			"PlanName":        plan_name,
			"AmountPaid":      payment_detail.Price,
			"Currency":        payment_detail.Currency,
			"PaymentMethod":   payment_detail.Payment_Method,
			"Date":            payment_detail.Start_Date,
			"TransactionDate": app.formatDate(transactionDate),
			"GrandTotal":      payment_detail.Price,
		}
		err = app.mailer.Send(user_email, "subscription_reciept.tmpl", data)
		if err != nil {
			app.logger.PrintError(err, nil)
		}
	})
	return nil
}

// createFailedTransactionHandler() A mono handler that created and saves a failed transaction
func (app *application) createFailedTransactionHandler(paymentDetails *data.Payment_Details, message, reference, error_code string) error {
	failedTransactionID, err := app.models.Payments.CreateFailedTransaction(paymentDetails, message, reference, error_code)
	if err != nil {
		return err
	}
	app.logger.PrintInfo("failed transaction", map[string]string{"Status": paymentDetails.Status, "id": fmt.Sprintf("%d", failedTransactionID)})
	return nil
}

// createChallengedTransaction() handler creates a challenged transaction in our database
// and sends a challange email to the user.
func (app *application) createChallengedTransaction(subscription *data.RecurringSubscription, VerifyResponse *data.VerifyResponse) error {
	err := app.models.Payments.CreateChallangedTransaction(subscription,
		VerifyResponse.Data.Authorization_url,
		*VerifyResponse.Data.Message,
		VerifyResponse.Data.Reference)

	if err != nil {
		return err
	}
	// send challange email to user notifying them of the challange
	app.background(func() {
		data := map[string]any{
			"UserName":             subscription.User_Name,
			"TransactionReference": VerifyResponse.Data.Reference,
			"PlanName":             subscription.Subscription.Plan_ID,
			"AmountPaid":           subscription.Subscription.Price,
			"Currency":             subscription.Currency,
			"PaymentMethod":        "card",
			"Date":                 app.formatDate(time.Now().UTC().String()),
			"TransactionDate":      app.formatDate(VerifyResponse.Data.TransactionDate),
			"GrandTotal":           VerifyResponse.Data.Amount,
		}
		err = app.mailer.Send(subscription.User_Email, "challanged_transaction.tmpl", data)
		if err != nil {
			app.logger.PrintError(err, nil)
		}
	})
	// report a nil error and proceed
	return nil
}

// updateExpiredSubscriptionHandler() is a bulk job, that is, everytime it's called it will
// update all expired subscriptions at once and not one by one.
// It will get back a slice of all subscriptions that have been updated. We will just log
// each of them out
func (app *application) updateExpiredSubscriptionHandler() {
	app.logger.PrintInfo("Updating expired subscriptions", nil)
	expiredSubscriptions, err := app.models.Payments.UpdateSubscriptionStatusAfterExpiration()
	if err != nil {
		app.logger.PrintError(err, nil)
	}
	for _, subscription := range expiredSubscriptions {
		app.logger.PrintInfo("Updated expired subscription", map[string]string{"Subscription ID": subscription.ID.String(),
			"User ID":    fmt.Sprintf("%d", subscription.User_ID),
			"Updated At": subscription.Updated_At.String()})
	}
}
