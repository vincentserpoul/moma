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
	Defi        string
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
