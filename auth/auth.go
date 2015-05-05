package auth

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/sauerbraten/persona"
)

type PersonaAuth struct {
	Users      AuthUsers
	PersonaUrl string
}

// Users should be a database table or something when using persona in production
type AuthUsers struct {
	Normals map[string]bool
	Admins  map[string]bool
}

// Config loader from json file
func LoadAuthUsers(fileName string) (AuthUsers, error) {
	var users AuthUsers

	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		return AuthUsers{}, err
	}

	if err := json.Unmarshal(file, &users); err != nil {
		return AuthUsers{}, err
	}

	return users, nil

}

// signs the client in by checking with the persona verification API and setting a secure session cookie.
// adds new users to the list of known users.
// passes the persona verifiation API response down to the client so the javascript can act on it.
func (pa *PersonaAuth) SignIn(resp http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	enc := json.NewEncoder(resp)

	response, err := persona.VerifyAssertion(pa.PersonaUrl, req.FormValue("assertion"))
	if err != nil {
		log.Println("sign in :", response.Email, " failed after persona VerifyAssertion")
		log.Println(err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	if response.OK() {
		setSessionCookie(resp, response.Email, response.Expires)

		if !pa.userExists(response.Email) {
			resp.WriteHeader(http.StatusUnauthorized)
			return
		}

		resp.WriteHeader(http.StatusOK)
		log.Println("sign in :", response.Email)
	} else {
		log.Println("sign in :", response.Email, " Response NOK from persona VerifyAssertion")
		resp.WriteHeader(http.StatusUnauthorized)
	}

	enc.Encode(response)
}

// revokes the cookie → client is signed out
func (pa *PersonaAuth) SignOut(resp http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	revokeSessionCookie(resp)
	resp.WriteHeader(http.StatusOK)
}

// Check if user exists against user list
func (pa *PersonaAuth) userExists(email string) bool {
	return pa.Users.Normals[email]
}

// Check if user exists against user list
func (pa *PersonaAuth) userIsAdmin(email string) bool {
	return pa.Users.Admins[email]
}

func (pa *PersonaAuth) IsAdmin() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			email := GetEmail(req)
			if email == "" {
				http.Error(res, "Not Authorized", http.StatusUnauthorized)
				return
			}
			isAdmin := pa.userIsAdmin(email)
			if !isAdmin {
				log.Fatal(email, " is not admin")
				http.Error(res, "Not Authorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(res, req)
		})
	}
}

// Basic returns a Handler that checks if user is logged in. Writes a http.StatusUnauthorized
// if not logged in
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
