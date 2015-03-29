package event

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/nu7hatch/gouuid"
)

// Represents a Physical Point in geographic notation [lat, lng].
type Point struct {
	Lat float64
	Lng float64
}

type Event struct {
	Id    string
	Email string
	Who   string
	When  time.Time
	Where Point
	What  string
	Pic   []string
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
