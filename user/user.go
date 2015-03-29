package user

import (
	"encoding/json"

	"github.com/garyburd/redigo/redis"
)

type User struct {
	Email     string
	EventList []string
}

func GetUserByEmail(redisco redis.Conn, email string) (User, error) {
	var usr User

	rawUsrByte, err := redisco.Do("GET", email)
	if err != nil {
		return usr, err
	}

	rawUsr, _ := redis.Bytes(rawUsrByte, nil)

	if rawUsr == nil {
		usr.Email = email
		usr.Save(redisco)
		return usr, nil
	}

	if err := json.Unmarshal(rawUsr, &usr); err != nil {
		return usr, err
	}

	return usr, nil

}

func (u *User) Save(redisco redis.Conn) error {

	user, err := json.Marshal(u)
	if err != nil {
		return err
	}

	redisco.Do("SET", u.Email, string(user))

	return nil
}
