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
)

type RedisHandler struct {
	RedisConn   redis.Conn
	personaAuth auth.PersonaAuth
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

	// Load user list, containing user profiles
	AuthUsers, err := auth.LoadAuthUsers(fmt.Sprintf("config/%s/auth.json", *env))
	if err != nil {
		log.Fatal(err)
	}

	// Associate this list to persona Auth
	var personaAuth auth.PersonaAuth
	personaAuth.Users = AuthUsers
	personaAuth.PersonaUrl = config.PersonaUrl

	// Put persona in handler
	red.personaAuth = personaAuth

	/* Public router */
	publicRouter := httprouter.New()

	/* Statics */
	publicRouter.ServeFiles("/js/*filepath", http.Dir("templates/js"))
	publicRouter.ServeFiles("/css/*filepath", http.Dir("templates/css"))
	publicRouter.ServeFiles("/img/*filepath", http.Dir("templates/img"))

	/* User log in */
	publicRouter.GET("/", red.User)
	publicRouter.POST("/auth/signin", personaAuth.SignIn)
	publicRouter.POST("/auth/signout", personaAuth.SignOut)

	middle := interpose.New()
	middle.UseHandler(publicRouter)

	/* Protected router */
	protectedRouter := httprouter.New()

	/* Save event */
	protectedRouter.POST("/events", red.SaveEvent)
	protectedRouter.DELETE("/events/:eventId", red.DeleteEvent)

	protectedMiddlew := interpose.New()
	protectedMiddlew.Use(auth.Persona())

	protectedMiddlew.UseHandler(protectedRouter)

	publicRouter.Handler("POST", "/events", protectedMiddlew)
	publicRouter.Handler("DELETE", "/events/:eventId", protectedMiddlew)

	/* Admin */

	/* admin router */
	adminRouter := httprouter.New()
	adminRouter.GET("/users", red.AdminUsers)
	adminMiddlew := interpose.New()
	adminMiddlew.Use(auth.Persona())
	adminMiddlew.Use(personaAuth.IsAdmin())

	adminMiddlew.UseHandler(adminRouter)

	publicRouter.Handler("GET", "/users", adminMiddlew)

	log.Fatal(http.ListenAndServe(config.Port, publicRouter))
}

type UserTemplate struct {
	User          user.User
	UserEventList []event.Event
}

func (red *RedisHandler) User(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	var UsrTpl UserTemplate

	// Check if User is authenticated
	email := auth.GetEmail(r)

	// If he is not
	if email == "" {
		err := templates.ExecuteTemplate(w, "index_signedout", nil)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	// If he is, we get his desc
	usr, err := user.GetUserByEmail(red.RedisConn, email)
	if err != nil {
		log.Fatal(err)
	}

	var userEvtLst []event.Event
	for _, eventId := range usr.EventList {
		evt, err := event.GetEventById(red.RedisConn, eventId)
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

func (red *RedisHandler) SaveEvent(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	email := auth.GetEmail(r)
	decoder := json.NewDecoder(r.Body)
	var evt event.Event
	err := decoder.Decode(&evt)
	if err != nil {
		panic(err)
	}
	evt.Email = email
	evt.Save(red.RedisConn)

	// Update user event list
	usr, err := user.GetUserByEmail(red.RedisConn, email)
	usr.EventList = append(usr.EventList, evt.Id)
	usr.Save(red.RedisConn)

	// return saved event
	response, _ := json.Marshal(evt)
	w.Write(response)

}

func (red *RedisHandler) DeleteEvent(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	email := auth.GetEmail(r)
	eventId := ps.ByName("eventId")
	// Update user event list
	usr, err := user.GetUserByEmail(red.RedisConn, email)
	if err != nil {
		log.Fatal(err)
		return
	}
	for i, userEventId := range usr.EventList {
		if userEventId == eventId {
			usr.EventList = append(usr.EventList[:i], usr.EventList[i+1:]...)
			break
		}
	}
	usr.Save(red.RedisConn)

	err = event.Delete(red.RedisConn, eventId)
	if err != nil {
		log.Fatal(err)
		return
	}
	w.Write([]byte("deleted"))
}

type AdminTemplate struct {
	Users          []user.User
	UsersEventList map[string][]event.Event
}

func (red *RedisHandler) AdminUsers(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	var AdminTpl AdminTemplate

	// We get all users
	for userEmail, _ := range red.personaAuth.Users.Normals {
		usr, _ := user.GetUserByEmail(red.RedisConn, userEmail)
		AdminTpl.Users = append(AdminTpl.Users, usr)
		var userEvtLst []event.Event
		for _, eventId := range usr.EventList {
			evt, err := event.GetEventById(red.RedisConn, eventId)
			if err != nil {
				log.Fatal(err)
			}
			userEvtLst = append(userEvtLst, evt)
		}
		AdminTpl.UsersEventList[usr.Email] = userEvtLst
	}

	err := templates.ExecuteTemplate(w, "admin_users", AdminTpl)
	if err != nil {
		log.Fatal(err)
	}
}
