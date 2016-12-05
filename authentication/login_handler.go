package authentication

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"github.com/fitstar/falcore"
	// "github.com/stuphlabs/pullcord"
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

func internalServerError(request *falcore.Request) *http.Response {
	return falcore.StringResponse(
		request.HttpRequest,
		500,
		nil,
		"<html><body><h1>Internal Server Error</h1>" +
		"<p>An internal server error occured." +
		"Please contact your system administrator.</p></body></html>",
	)
}

func notImplementedError(request *falcore.Request) *http.Response {
	return falcore.StringResponse(
		request.HttpRequest,
		501,
		nil,
		"<html><body><h1>Not Implemented</h1>" +
		"<p>The requested behavior has not yet been implemented." +
		"Please contact your system administrator.</p></body></html>",
	)
}

func loginPage(
	request *falcore.Request,
	sesh Session,
	badCreds bool,
) *http.Response {
	return falcore.StringResponse(
		request.HttpRequest,
		501,
		nil,
		"<html><body><h1>Not Implemented</h1>" +
		"<p>The requested behavior has not yet been implemented." +
		"Please contact your system administrator.</p></body></html>",
	)
}

func (handler *LoginHandler) handleLogin(
	request *falcore.Request,
) *http.Response {
	errString := ""
	rawsesh, present := request.Context["session"]
	if !present {
		log().Crit(
			"login handler was unable to retrieve session from" +
			" context",
		)
		return internalServerError(request)
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
		return internalServerError(request)
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
		return internalServerError(request)
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
		return internalServerError(request)
	} else if err = sesh.SetValue(authSeshKey, true); err != nil {
		log().Err(
			fmt.Sprintf(
				"login handler error during auth set: %v",
				err,
			),
		)
		return internalServerError(request)
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
		return internalServerError(request)
	}
	nextXsrfToken := hex.EncodeToString(rawXsrfToken)

	if err = sesh.SetValue(xsrfKey, nextXsrfToken); err != nil {
		log().Err(
			fmt.Sprintf(
				"login handler error during xsrf set: %v",
				err,
			),
		)
		return internalServerError(request)
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

// NewLoginFilter generates a Falcore RequestFilter that presents a login page
// to unauthenticated users while getting out of the way as quickly as possible
// for authenticated users. While no requests of any kind will make it past
// this layer in the event of an unauthenticated user, a request from an
// authenticated user will be passed along with the only modifications being an
// attempt to remove any trace that this layer existed. So long as the
// LoginHandler's identifier is unique, requests being passed through this
// login filter should not interfere in any way with the behavior of a
// third-party web app that exists downstream from this filter.
func NewLoginFilter(
	sessionHandler SessionHandler,
	loginHandler LoginHandler,
) falcore.RequestFilter {
	log().Info("initializing login handler")

	return NewCookiemaskFilter(
		sessionHandler,
		falcore.NewRequestFilter(loginHandler.handleLogin),
		falcore.NewRequestFilter(
			func (request *falcore.Request) *http.Response {
				return internalServerError(request)
			},
		),
	)
}

