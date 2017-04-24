package authentication

import (
	"bytes"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/fitstar/falcore"
	// "github.com/stuphlabs/pullcord"
	"github.com/stuphlabs/pullcord/config"
	"github.com/stuphlabs/pullcord/util"
	"net/http"
)

const XsrfTokenLength = 64

// LoginHandler is a login handling system that presents a login page backed by
// a PasswordChecker for users that are not yet logged in, while seamlessly
// forwarding all requests downstream for users that are logged in. A
// LoginHandler has an identifier (which it uses to differentiate its login
// tokens and authentication flags from other components, possibly including
// other LoginHandlers), a PasswordChecker (which it allows users to
// authenticate against in conjunction with its own XSRF token), and a
// downstream RequestFilter (possibly an entire pipeline).
type LoginHandler struct {
	Identifier string
	PasswordChecker PasswordChecker
	Downstream falcore.RequestFilter
}

func init() {
	config.RegisterResourceType(
		"loginhandler",
		func() json.Unmarshaler {
			return new(LoginHandler)
		},
	)
}

func (h *LoginHandler) UnmarshalJSON(input []byte) (error) {
	var t struct {
		Identifier string
		PasswordChecker config.Resource
		Downstream config.Resource
	}

	dec := json.NewDecoder(bytes.NewReader(input))
	if e := dec.Decode(&t); e != nil {
		log().Err("Unable to decode LoginHandler")
		return e
	} else {
		p := t.PasswordChecker.Unmarshaled
		switch p := p.(type) {
		case PasswordChecker:
			h.PasswordChecker = p
		default:
			log().Err(
				fmt.Sprintf(
					"Registry value is not a" +
					" PasswordChecker: %s",
					t.PasswordChecker,
				),
			)
			return config.UnexpectedResourceType
		}

		d := t.Downstream.Unmarshaled
		switch d := d.(type) {
		case falcore.RequestFilter:
			h.Downstream = d
		default:
			log().Err(
				fmt.Sprintf(
					"Registry value is not a" +
					" RequestFilter: %s",
					t.Downstream,
				),
			)
			return config.UnexpectedResourceType
		}

		h.Identifier = t.Identifier

		return nil
	}
}

func (handler *LoginHandler) FilterRequest(
	request *falcore.Request,
) *http.Response {
	errString := ""
	rawsesh, present := request.Context["session"]
	if !present {
		log().Crit(
			"login handler was unable to retrieve session from" +
			" context",
		)
		return util.InternalServerError.FilterRequest(request)
	}
	sesh := rawsesh.(Session)

	authSeshKey := "authenticated-" + handler.Identifier
	xsrfKey := "xsrf-" + handler.Identifier
	usernameKey := "username-" + handler.Identifier
	passwordKey := "password-" + handler.Identifier

	authd, err := sesh.GetValue(authSeshKey)
	if err == nil && authd == true {
		log().Debug("login handler passing request along")
		return handler.Downstream.FilterRequest(request)
	} else if err != NoSuchSessionValueError {
		log().Err(
			fmt.Sprintf(
				"login handler error during auth status" +
				" retrieval: %v",
				err,
			),
		)
		return util.InternalServerError.FilterRequest(request)
	}

	xsrfStored, err := sesh.GetValue(xsrfKey)
	if err != nil && err != NoSuchSessionValueError {
		log().Err(
			fmt.Sprintf(
				"login handler error during xsrf token" +
				" retrieval: %v",
				err,
			),
		)
		return util.InternalServerError.FilterRequest(request)
	} else if err == NoSuchSessionValueError {
		log().Info("login handler received new request")
	} else if err = request.HttpRequest.ParseForm(); err != nil {
		log().Warning(
			fmt.Sprintf(
				"login handler error during ParseForm: %v",
				err,
			),
		)
		errString = "Bad request"
	} else if xsrfRcvd, present :=
		request.HttpRequest.PostForm[xsrfKey]; ! present {
		log().Info("login handler did not receive xsrf token")
		errString = "Invalid credentials"
	} else  if len(xsrfRcvd) != 1 || 1 != subtle.ConstantTimeCompare(
		[]byte(xsrfStored.(string)),
		[]byte(xsrfRcvd[0]),
	) {
		log().Info("login handler received bad xsrf token")
		errString = "Invalid credentials"
	} else if uVals, present :=
		request.HttpRequest.PostForm[usernameKey]; ! present {
		log().Info("login handler did not receive username")
		errString = "Invalid credentials"
	} else if pVals, present :=
		request.HttpRequest.PostForm[passwordKey]; ! present {
		log().Info("login handler did not receive password")
		errString = "Invalid credentials"
	} else if len(uVals) != 1 || len(pVals) != 1 {
		log().Info(
			"login handler received multi values for username or" +
			" password",
		)
		errString = "Bad request"
	} else if err = handler.PasswordChecker.CheckPassword(
		uVals[0],
		pVals[0],
	); err == NoSuchIdentifierError {
		log().Info("login handler received bad username")
		errString = "Invalid credentials"
	} else if err == BadPasswordError {
		log().Info("login handler received bad password")
		errString = "Invalid credentials"
	} else if err != nil{
		log().Err(
			fmt.Sprintf(
				"login handler error during CheckPassword: %v",
				err,
			),
		)
		return util.InternalServerError.FilterRequest(request)
	} else if err = sesh.SetValue(authSeshKey, true); err != nil {
		log().Err(
			fmt.Sprintf(
				"login handler error during auth set: %v",
				err,
			),
		)
		return util.InternalServerError.FilterRequest(request)
	} else {
		log().Notice(
			fmt.Sprintf(
				"login successful for: %s",
				uVals[0],
			),
		)
		return handler.Downstream.FilterRequest(request)
	}

	rawXsrfToken := make([]byte, XsrfTokenLength)
	if rsize, err := rand.Read(
		rawXsrfToken[:],
	); err != nil || rsize != XsrfTokenLength {
		log().Err(
			fmt.Sprintf(
				"login handler error during xsrf generation:" +
				" len expected: %u, actual: %u, err: %v",
				XsrfTokenLength,
				rsize,
				err,
			),
		)
		return util.InternalServerError.FilterRequest(request)
	}
	nextXsrfToken := hex.EncodeToString(rawXsrfToken)

	if err = sesh.SetValue(xsrfKey, nextXsrfToken); err != nil {
		log().Err(
			fmt.Sprintf(
				"login handler error during xsrf set: %v",
				err,
			),
		)
		return util.InternalServerError.FilterRequest(request)
	}

	errMarkup := ""
	if errString != "" {
		errMarkup = fmt.Sprintf(
			"<label class=\"error\">%s</label>",
			errString,
		)
	}

	return falcore.StringResponse(
		request.HttpRequest,
		200,
		nil,
		fmt.Sprintf(
			"<html><head><title>Pullcord Login</title></head>" +
			"<body><form method=\"POST\" action=\"%s\">" +
			"<fieldset><legend>Pullcord Login</legend>%s<label " +
			"for=\"username\">Username:</label><input " +
			"type=\"text\" name=\"%s\" id=\"username\" /><label " +
			"for=\"password\">Password:</label><input " +
			"type=\"password\" name=\"%s\" id=\"password\" />" +
			"<input type=\"hidden\" name=\"%s\" value=\"%s\" />" +
			"<input type=\"submit\" value=\"Login\"/></fieldset>" +
			"</form></body></html>",
			request.HttpRequest.URL.Path,
			errMarkup,
			usernameKey,
			passwordKey,
			xsrfKey,
			nextXsrfToken,
		),
	)
}

