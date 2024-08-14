package main

import (
	"net/http"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	//time.Sleep(5 * time.Second)
	err := app.writeJSON(w, http.StatusOK, app.returnEnvInfo(), nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}
