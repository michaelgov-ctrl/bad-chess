package main

import (
	"net/http"

	"github.com/michaelgov-ctrl/bad-chess/internal/validator"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	app.render(w, r, http.StatusOK, "home.tmpl.html", data)
}

func (app *application) matchMakingHandler(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	app.render(w, r, http.StatusOK, "matchmaking.tmpl.html", data)
}

func (app *application) matchesHandler(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	app.render(w, r, http.StatusOK, "match.tmpl.html", data)
}

type userLoginForm struct {
	Key                 string `form:"key"`
	validator.Validator `form:"-"`
}

func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userLoginForm{}
	app.render(w, r, http.StatusOK, "login.tmpl.html", data)
}

func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) {
	var form userLoginForm
	if err := app.decodePostForm(r, &form); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Key), "key", "This field cannot be blank")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "login.tmpl.html", data)
		return
	}

	id, err := app.authentication.Authenticate(form.Key)
	if err != nil {
		// TODO: this will have to change to the forwarded for once behind a proxy
		app.logger.Info("failed authentication attempt", "origin", r.RemoteAddr)
		form.AddNonFieldError("key is incorrect")
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "login.tmpl.html", data)
		return
	}

	if err := app.sessionManager.RenewToken(r.Context()); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), authenticatedSessionKeyName, id)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	if err := app.sessionManager.RenewToken(r.Context()); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Remove(r.Context(), authenticatedSessionKeyName)
	app.sessionManager.Put(r.Context(), "flash", "You've been logged out")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
