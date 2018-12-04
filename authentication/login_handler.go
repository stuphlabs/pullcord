package authentication

import (
	"bytes"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/proidiot/gone/log"
	"github.com/stuphlabs/pullcord/config"
	"github.com/stuphlabs/pullcord/util"
)

// XsrfTokenLength is the length of XSRF token strings.
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
	Identifier      string
	PasswordChecker PasswordChecker
	Downstream      http.Handler
}

func init() {
	config.RegisterResourceType(
		"loginhandler",
		func() json.Unmarshaler {
			return new(LoginHandler)
		},
	)
}

// UnmarshalJSON implements encoding/json.Unmarshaler.
func (h *LoginHandler) UnmarshalJSON(input []byte) error {
	var t struct {
		Identifier      string
		PasswordChecker config.Resource
		Downstream      config.Resource
	}

	dec := json.NewDecoder(bytes.NewReader(input))
	if e := dec.Decode(&t); e != nil {
		log.Err("Unable to decode LoginHandler")
		return e
	} else {
		p := t.PasswordChecker.Unmarshaled
		switch p := p.(type) {
		case PasswordChecker:
			h.PasswordChecker = p
		default:
			log.Err(
				fmt.Sprintf(
					"Registry value is not a"+
						" PasswordChecker: %#v",
					t.PasswordChecker,
				),
			)
			return config.UnexpectedResourceType
		}

		if d, ok := t.Downstream.Unmarshaled.(http.Handler); ok {
			h.Downstream = d
		} else {
			log.Err(
				fmt.Sprintf(
					"Registry value is not a"+
						" RequestFilter: %#v",
					t.Downstream,
				),
			)
			return config.UnexpectedResourceType
		}

		h.Identifier = t.Identifier

		return nil
	}
}

func (handler *LoginHandler) ServeHTTP(
	w http.ResponseWriter,
	request *http.Request,
) {
	errString := ""
	rawsesh := request.Context().Value("session")
	if rawsesh == nil {
		log.Crit(
			"login handler was unable to retrieve session from" +
				" context",
		)
		util.InternalServerError.ServeHTTP(w, request)
		return
	}
	sesh := rawsesh.(Session)

	authSeshKey := "authenticated-" + handler.Identifier
	xsrfKey := "xsrf-" + handler.Identifier
	usernameKey := "username-" + handler.Identifier
	passwordKey := "password-" + handler.Identifier

	authd, err := sesh.GetValue(authSeshKey)
	if err == nil && authd == true {
		log.Debug("login handler passing request along")
		handler.Downstream.ServeHTTP(w, request)
		return
	} else if err != NoSuchSessionValueError {
		log.Err(
			fmt.Sprintf(
				"login handler error during auth status"+
					" retrieval: %v",
				err,
			),
		)
		util.InternalServerError.ServeHTTP(w, request)
		return
	}

	xsrfStored, err := sesh.GetValue(xsrfKey)
	if err != nil && err != NoSuchSessionValueError {
		log.Err(
			fmt.Sprintf(
				"login handler error during xsrf token"+
					" retrieval: %v",
				err,
			),
		)
		util.InternalServerError.ServeHTTP(w, request)
		return
	} else if err == NoSuchSessionValueError {
		log.Info("login handler received new request")
	} else if err = request.ParseForm(); err != nil {
		log.Warning(
			fmt.Sprintf(
				"login handler error during ParseForm: %#v",
				err,
			),
		)
		errString = "Bad request"
	} else if xsrfRcvd, present :=
		request.PostForm[xsrfKey]; !present {
		log.Info("login handler did not receive xsrf token")
		errString = "Invalid credentials"
	} else if len(xsrfRcvd) != 1 || 1 != subtle.ConstantTimeCompare(
		[]byte(xsrfStored.(string)),
		[]byte(xsrfRcvd[0]),
	) {
		log.Info("login handler received bad xsrf token")
		errString = "Invalid credentials"
	} else if uVals, present :=
		request.PostForm[usernameKey]; !present {
		log.Info("login handler did not receive username")
		errString = "Invalid credentials"
	} else if pVals, present :=
		request.PostForm[passwordKey]; !present {
		log.Info("login handler did not receive password")
		errString = "Invalid credentials"
	} else if len(uVals) != 1 || len(pVals) != 1 {
		log.Info(
			"login handler received multi values for username or" +
				" password",
		)
		errString = "Bad request"
	} else if err = handler.PasswordChecker.CheckPassword(
		uVals[0],
		pVals[0],
	); err == NoSuchIdentifierError {
		log.Info("login handler received bad username")
		errString = "Invalid credentials"
	} else if err == BadPasswordError {
		log.Info("login handler received bad password")
		errString = "Invalid credentials"
	} else if err != nil {
		log.Err(
			fmt.Sprintf(
				"login handler error during CheckPassword: %#v",
				err,
			),
		)
		util.InternalServerError.ServeHTTP(w, request)
		return
	} else if err = sesh.SetValue(authSeshKey, true); err != nil {
		log.Err(
			fmt.Sprintf(
				"login handler error during auth set: %#v",
				err,
			),
		)
		util.InternalServerError.ServeHTTP(w, request)
		return
	} else {
		log.Notice(
			fmt.Sprintf(
				"login successful for: %s",
				uVals[0],
			),
		)
		handler.Downstream.ServeHTTP(w, request)
		return
	}

	rawXsrfToken := make([]byte, XsrfTokenLength)
	if rsize, err := rand.Read(
		rawXsrfToken[:],
	); err != nil || rsize != XsrfTokenLength {
		log.Err(
			fmt.Sprintf(
				"login handler error during xsrf generation:"+
					" len expected: %d, actual: %d,"+
					" err: %#v",
				XsrfTokenLength,
				rsize,
				err,
			),
		)
		util.InternalServerError.ServeHTTP(w, request)
		return
	}
	nextXsrfToken := hex.EncodeToString(rawXsrfToken)

	if err = sesh.SetValue(xsrfKey, nextXsrfToken); err != nil {
		log.Err(
			fmt.Sprintf(
				"login handler error during xsrf set: %#v",
				err,
			),
		)
		util.InternalServerError.ServeHTTP(w, request)
		return
	}

	errMarkup := ""
	if errString != "" {
		errMarkup = fmt.Sprintf(
			"<label class=\"error\">%s</label><br />",
			errString,
		)
	}

	fmt.Fprintf(
		w,
		"<html><head><title>Pullcord Login</title></head><body>"+
			"<form method=\"POST\" action=\"%s\"><fieldset>"+
			"<legend>Pullcord Login</legend>%s"+
			"<label for=\"username\">Username:</label>"+
			"<input type=\"text\" name=\"%s\" id=\"username\" />"+
			"<label for=\"password\">Password:</label>"+
			"<input type=\"password\" name=\"%s\""+
			"id=\"password\" /><input type=\"hidden\" name=\"%s\""+
			" value=\"%s\" /><input type=\"submit\""+
			" value=\"Login\"/></fieldset></form></body></html>",
		request.URL.Path,
		errMarkup,
		usernameKey,
		passwordKey,
		xsrfKey,
		nextXsrfToken,
	)
	return
}
