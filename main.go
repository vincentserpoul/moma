package main /* import "github.com/vincentserpoul/moma" */

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"html/template"
	"net/http"

	"github.com/carbocation/interpose"
	"github.com/garyburd/redigo/redis"
	"github.com/julienschmidt/httprouter"
	"github.com/vincentserpoul/moma/auth"
	"github.com/vincentserpoul/moma/event"
	"github.com/vincentserpoul/moma/user"
	"github.com/vincentserpoul/moma/utils"
	"github.com/vincentserpoul/rbac"
	rbacjson "github.com/vincentserpoul/rbac/json"
)

// RedisHandler includes all necessary info for Pages
type RedisHandler struct {
	RedisConn   redis.Conn
	personaAuth auth.PersonaAuth
	authoriz    rbacjson.RbacConf
}

var templates = template.Must(template.ParseGlob("templates/*.go.html"))

func main() {

	// flags
	env := flag.String("env", "prod", "environment, if none specified, it will prod")
	flag.Parse()

	// Load config file, containing redis conf and app conf
	config, err := utils.LoadConfig(fmt.Sprintf("config/%s/app.json", *env))
	if err != nil {
		log.Fatal(err)
	}

	// Connect to redis
	redisco, err := redis.Dial("tcp", fmt.Sprintf("%s%s", config.Redis.Host, config.Redis.Port))
	if err != nil {
		log.Fatal(err)
	}
	defer redisco.Close()

	var red RedisHandler
	red.RedisConn = redisco

	// Associate this list to persona Auth
	var personaAuth auth.PersonaAuth
	personaAuth.PersonaURL = config.PersonaURL

	// Put persona in handler
	red.personaAuth = personaAuth

	/* Public router */
	publicRouter := httprouter.New()

	/* Statics */
	publicRouter.ServeFiles("/js/*filepath", http.Dir("templates/js"))
	publicRouter.ServeFiles("/css/*filepath", http.Dir("templates/css"))
	publicRouter.ServeFiles("/img/*filepath", http.Dir("templates/img"))

	/* User log in */
	publicRouter.GET("/", Landing)
	publicRouter.POST("/auth/signin", personaAuth.SignIn)
	publicRouter.POST("/auth/signout", personaAuth.SignOut)

	middle := interpose.New()
	middle.UseHandler(publicRouter)

	/* Protected router */
	protectedRouter := httprouter.New()

	/* Save event */
	protectedRouter.GET("/user", red.User)
	protectedRouter.POST("/events", red.SaveEvent)
	protectedRouter.DELETE("/events/:eventId", red.DeleteEvent)
	protectedRouter.PUT("/events/:eventId", red.UpdateEvent)
	protectedRouter.GET("/admin", red.Admin)
	protectedRouter.GET("/admin/eventlist", red.AdminEvent)

	protectedMiddlew := interpose.New()
	protectedMiddlew.Use(auth.Persona())

	// Init RBAC with files
	var authoriz rbacjson.RbacConf
	authoriz.AppRoleActions, err = rbacjson.LoadRoleActionsFromFile(fmt.Sprintf("config/%s/roleactions.json", *env))
	if err != nil {
		log.Fatal(err)
	}

	authoriz.AppUserRole, err = rbacjson.LoadUserRoleFromFile(fmt.Sprintf("config/%s/userrole.json", *env))
	if err != nil {
		log.Fatal(err)
	}

	//Saving it for future use in RedisHandler
	red.authoriz = authoriz

	funcGetUserID := auth.GetEmail
	protectedMiddlew.Use(rbac.InterposeRBAC(authoriz, funcGetUserID))

	protectedMiddlew.UseHandler(protectedRouter)

	publicRouter.Handler("GET", "/user", protectedMiddlew)
	publicRouter.Handler("POST", "/events", protectedMiddlew)
	publicRouter.Handler("PUT", "/events/:eventId", protectedMiddlew)
	publicRouter.Handler("DELETE", "/events/:eventId", protectedMiddlew)
	publicRouter.Handler("GET", "/admin", protectedMiddlew)
	publicRouter.Handler("GET", "/admin/eventlist", protectedMiddlew)

	log.Fatal(http.ListenAndServe(config.Port, publicRouter))
}

// Landing is the simple home page handler
func Landing(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	email := auth.GetEmail(r)
	if email != "" {
		log.Println("logging in ", email)
		http.Redirect(w, r, "http://"+r.Host+"/user", http.StatusFound)
	}

	err := templates.ExecuteTemplate(w, "index_signedout", 302)
	if err != nil {
		log.Fatal(err)
	}
	return
}

// UserTemplate is going to be used foreach user
type UserTemplate struct {
	User          user.User
	UserEventList []event.Event
}

// User is handling the get request to the main user page
func (red *RedisHandler) User(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	var UsrTpl UserTemplate

	// Check if User is authenticated
	email := auth.GetEmail(r)

	// If he is, we get his desc
	usr, err := user.GetUserByEmail(red.RedisConn, email)
	if err != nil {
		log.Fatal(err)
	}

	var userEvtLst []event.Event
	for _, eventID := range usr.EventList {
		evt, err := event.GetEventById(red.RedisConn, eventID)
		if err != nil {
			log.Fatal(err)
		}
		userEvtLst = append(userEvtLst, evt)
	}

	UsrTpl.User = usr
	UsrTpl.UserEventList = userEvtLst
	err = templates.ExecuteTemplate(w, "index_signedin", UsrTpl)
	if err != nil {
		log.Fatal(err)
	}
}

// SaveEvent is handling the POST event
func (red *RedisHandler) SaveEvent(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	email := auth.GetEmail(r)
	decoder := json.NewDecoder(r.Body)
	var evt event.Event
	err := decoder.Decode(&evt)

	evt.Email = email
	err = evt.Save(red.RedisConn)

	// Update user event list
	usr, err := user.GetUserByEmail(red.RedisConn, email)
	usr.EventList = append(usr.EventList, evt.Id)
	err = usr.Save(red.RedisConn)
	if err != nil {
		panic(err)
	}
	// return saved event
	response, _ := json.Marshal(evt)
	w.Write(response)

}

// DeleteEvent is handling the delete of the event
func (red *RedisHandler) DeleteEvent(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	email := auth.GetEmail(r)
	eventID := ps.ByName("eventId")
	// Update user event list
	usr, err := user.GetUserByEmail(red.RedisConn, email)
	if err != nil {
		log.Fatal(err)
		return
	}
	for i, userEventID := range usr.EventList {
		if userEventID == eventID {
			usr.EventList = append(usr.EventList[:i], usr.EventList[i+1:]...)
			break
		}
	}
	err = usr.Save(red.RedisConn)

	err = event.Delete(red.RedisConn, eventID)
	if err != nil {
		log.Fatal(err)
		return
	}
	w.Write([]byte("deleted"))
}

// AdminTemplate is going to be used foreach user
type AdminTemplate struct {
	User  user.User
	Users []UserTemplate
}

// Admin is handling the get request to the main user page
func (red *RedisHandler) Admin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	var AdmTpl AdminTemplate
	var err error

	AdmTpl.User, err = user.GetUserByEmail(red.RedisConn, auth.GetEmail(r))
	if err != nil {
		log.Fatal(err)
	}

	for _, userRole := range red.authoriz.AppUserRole {
		usrTmpl := new(UserTemplate)
		// If he is, we get his desc
		usr, err := user.GetUserByEmail(red.RedisConn, userRole.UserID)
		if err != nil {
			log.Fatal(err)
		}
		var userEvtLst []event.Event
		for _, eventID := range usr.EventList {
			evt, err := event.GetEventById(red.RedisConn, eventID)
			if err != nil {
				log.Fatal(err)
			}
			userEvtLst = append(userEvtLst, evt)
		}
		usrTmpl.User = usr
		usrTmpl.UserEventList = userEvtLst

		AdmTpl.Users = append(AdmTpl.Users, *usrTmpl)
		usrTmpl = nil
	}

	err = templates.ExecuteTemplate(w, "admin_users", AdmTpl)
	if err != nil {
		log.Fatal(err)
	}
}

// AdminTemplate is going to be used foreach user
type AdminEventTemplate struct {
	User                 user.User
	EventListTeam1       []event.Event
	EventListTeam2       []event.Event
	EventListNotAssigned []event.Event
}

// Admin is handling the get request to the main user page
func (red *RedisHandler) AdminEvent(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	var AdmTpl AdminEventTemplate
	var err error

	AdmTpl.User, err = user.GetUserByEmail(red.RedisConn, auth.GetEmail(r))
	if err != nil {
		log.Fatal(err)
	}

	var allEvtLst []event.Event

	for _, userRole := range red.authoriz.AppUserRole {
		// If he is, we get his desc
		usr, err := user.GetUserByEmail(red.RedisConn, userRole.UserID)
		if err != nil {
			log.Fatal(err)
		}

		for _, eventID := range usr.EventList {
			evt, err := event.GetEventById(red.RedisConn, eventID)
			if err != nil {
				log.Fatal(err)
			}
			allEvtLst = append(allEvtLst, evt)
		}
	}

	var nextTeam1Id = ""
	var nextTeam2Id = ""

	// reorder event list according to start bool and next
	// Find all the starting points
	for _, usrEvent := range allEvtLst {
		if usrEvent.StartTeam1 || nextTeam1Id == usrEvent.Id {
			AdmTpl.EventListTeam1 = append(AdmTpl.EventListTeam1, usrEvent)
			nextTeam1Id = usrEvent.NextEventId
		}
		if usrEvent.StartTeam2 || nextTeam2Id == usrEvent.Id {
			AdmTpl.EventListTeam2 = append(AdmTpl.EventListTeam2, usrEvent)
			nextTeam2Id = usrEvent.NextEventId
		}
	}

	var alreadyAssigned bool

	for _, usrEvent := range allEvtLst {
		alreadyAssigned = false
		for _, usrEventTeam1 := range AdmTpl.EventListTeam1 {
			if usrEventTeam1.Id == usrEvent.Id {
				alreadyAssigned = true
			}
		}
		for _, usrEventTeam2 := range AdmTpl.EventListTeam1 {
			if usrEventTeam2.Id == usrEvent.Id {
				alreadyAssigned = true
			}
		}

		if !alreadyAssigned {
			AdmTpl.EventListNotAssigned = append(AdmTpl.EventListNotAssigned, usrEvent)
		}
	}

	err = templates.ExecuteTemplate(w, "admin_event_chain", AdmTpl)
	if err != nil {
		log.Fatal(err)
	}
}

// SaveEvent is handling the POST event
func (red *RedisHandler) UpdateEvent(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	decoder := json.NewDecoder(r.Body)
	var evt event.Event
	err := decoder.Decode(&evt)

	err = evt.Save(red.RedisConn)

	if err != nil {
		log.Fatal(err)
	}
	// return saved event
	response, _ := json.Marshal(evt)
	w.Write(response)

}
