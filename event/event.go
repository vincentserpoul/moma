package event

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/nu7hatch/gouuid"
)

// Point represents a Physical Point in geographic notation [lat, lng].
type Point struct {
	Lat float64
	Lng float64
}

// Event used for events
type Event struct {
	Id          string
	Email       string
	Who         string
	When        time.Time
	Where       Point
	WhereApprox string
	What        string
	Pic         []string
	ShortURL    string
	QRCode      string
	StartTeam1  bool
	StartTeam2  bool
	NextEventId string
	NextPlaceQ  string
	CurrPlace   string
	Defi        string
	Rebus       string
}

func GetEventById(redisco redis.Conn, eventId string) (Event, error) {
	var evt Event

	rawEvtByte, err := redisco.Do("GET", eventId)
	if err != nil {
		return evt, err
	}

	rawEvt, _ := redis.Bytes(rawEvtByte, nil)

	if err := json.Unmarshal(rawEvt, &evt); err != nil {
		return evt, err
	}

	return evt, nil
}

func (e *Event) Save(redisco redis.Conn) error {
	// generate its Id if needed
	e.GenerateNewEventId()
	jsonEvt, err := json.Marshal(e)
	if err != nil {
		log.Fatal(err)
	}

	_, err = redisco.Do("SET", e.Id, string(jsonEvt))

	return err
}

func (e *Event) SaveJSON(jsonEvt string, redisco redis.Conn) error {
	_, err := redisco.Do("SET", e.Id, jsonEvt)

	return err
}

func (e *Event) GenerateNewEventId() error {
	if e.Id != "" {
		return errors.New("This event already has an id")
	}
	newId, err := uuid.NewV4()
	e.Id = newId.String()
	if err != nil {
		return err
	}

	return nil
}

func Delete(redisco redis.Conn, eventId string) error {
	_, err := redisco.Do("DEL", eventId)

	return err
}

func (e *Event) GetFirstImage() string {
	if len(e.Pic) > 0 {
		return e.Pic[0]
	}
	return ""
}

func (e *Event) GetFormattedDate() string {
	return e.When.Format("02/01/2006")
}

func (e *Event) GetFormattedDateForSave() string {
	return e.When.Format("2006/01/02")
}

func (e *Event) GetJSON() string {
	jsonEvt, _ := json.Marshal(e)

	return string(jsonEvt)
}

func (e *Event) GetShortWhat() string {
	shortEvt := e.What
	if len(e.What) > 100 {
		shortEvt = e.What[:100] + "..."
	}
	return shortEvt
}

func GetEventInEventList(id string, allEvtLst []*Event) *Event {
	// Find the top each list
	for _, usrEvent := range allEvtLst {
		if usrEvent.Id == id {
			return usrEvent
		}
	}

	return nil
}

func GetStartingEvent(team string, allEvtLst []*Event) *Event {
	// Find the top each list
	for _, usrEvent := range allEvtLst {
		if team == "team1" {
			if usrEvent.StartTeam1 {
				return usrEvent
			}
		}
		if team == "team2" {
			if usrEvent.StartTeam2 {
				return usrEvent
			}
		}
	}

	return nil
}

func GetTeamEvents(team string, allEvtLst []*Event) []*Event {
	startingEvent := GetStartingEvent(team, allEvtLst)
	var teamEventList []*Event
	var nextEvent *Event

	if startingEvent == nil {
		return teamEventList
	}
	teamEventList = append(teamEventList, startingEvent)

	nextEventId := startingEvent.NextEventId
	safeLoop := 0
	for nextEventId != "" {
		nextEvent = GetEventInEventList(nextEventId, allEvtLst)
		teamEventList = append(teamEventList, nextEvent)
		nextEventId = nextEvent.NextEventId
		safeLoop++
	}

	return teamEventList
}

func GetNoTeamEvents(team1EventList []*Event, team2EventList []*Event, allEvtLst []*Event) []*Event {
	var alreadyAssigned bool
	var noTeamEventList []*Event
	for _, usrEvent := range allEvtLst {
		alreadyAssigned = false
		for _, usrEventTeam1 := range team1EventList {
			if usrEventTeam1.Id == usrEvent.Id {
				alreadyAssigned = true
			}
		}
		for _, usrEventTeam2 := range team2EventList {
			if usrEventTeam2.Id == usrEvent.Id {
				alreadyAssigned = true
			}
		}

		if !alreadyAssigned {
			noTeamEventList = append(noTeamEventList, usrEvent)
		}
	}

	return noTeamEventList
}
