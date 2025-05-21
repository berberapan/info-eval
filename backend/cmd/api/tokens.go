package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/berberapan/info-eval/internal/data"
	"github.com/berberapan/info-eval/internal/validator"

	"github.com/pascaldekloe/jwt"
)

func (app *application) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	v := validator.New()
	data.ValidateEmail(v, input.Email)
	data.ValidatePlainTextPassword(v, input.Password)
	if !v.Valid() {
		app.failedValidateResponse(w, r, v.Errors)
		return
	}
	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if !match {
		app.invalidCredentialsResponse(w, r)
		return
	}
	var claims jwt.Claims
	claims.Subject = user.ID.String()
	claims.Issued = jwt.NewNumericTime(time.Now())
	claims.NotBefore = jwt.NewNumericTime(time.Now())
	claims.Expires = jwt.NewNumericTime(time.Now().Add(12 * time.Hour))
	claims.Issuer = "infoeval.boukdir.se"
	claims.Audiences = []string{"infoeval.boukdir.se"}

	jwtBytes, err := claims.HMACSign(jwt.HS256, []byte(app.config.jwt.secret))
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	cookie := http.Cookie{
		Name:     "token",
		Value:    string(jwtBytes),
		Path:     "/",
		Expires:  time.Now().Add(12 * time.Hour),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &cookie)

	err = app.writeJSON(w, http.StatusCreated, jsonEnvelope{"message": "authentication successful", "user_id": user.ID.String()}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) removeAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	cookie := http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &cookie)
	err := app.writeJSON(w, http.StatusOK, jsonEnvelope{"message": "authentication cleared"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
