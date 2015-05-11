package auth

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/sauerbraten/persona"
)

// PersonaAuth will contain the persona URL to return to
type PersonaAuth struct {
	PersonaURL string
}

// SignIn signs the client in by checking with the persona verification API and setting a secure session cookie.
// adds new users to the list of known users.
// passes the persona verifiation API response down to the client so the javascript can act on it.
func (pa *PersonaAuth) SignIn(resp http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	enc := json.NewEncoder(resp)

	response, err := persona.VerifyAssertion(pa.PersonaURL, req.FormValue("assertion"))
	if err != nil {
		log.Println("sign in :", response.Email, " failed after persona VerifyAssertion")
		log.Println(err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	if response.OK() {
		setSessionCookie(resp, response.Email, response.Expires)
		resp.WriteHeader(http.StatusOK)
		log.Println("sign in :", response.Email)
	} else {
		log.Println("sign in :", response.Email, " Response NOK from persona VerifyAssertion")
		resp.WriteHeader(http.StatusUnauthorized)
	}

	enc.Encode(response)
}

// SignOut revokes the cookie → client is signed out
func (pa *PersonaAuth) SignOut(resp http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	revokeSessionCookie(resp)
	resp.WriteHeader(http.StatusOK)
}

// Persona Basic returns a Handler that checks if user is logged in. Writes a http.StatusUnauthorized
func Persona() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			auth := GetEmail(req)
			if auth == "" {
				http.Error(res, "Not Authorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(res, req)
		})
	}
}
