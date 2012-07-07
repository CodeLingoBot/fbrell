package empcheck

import (
	"fmt"
	"github.com/daaku/go.fbapi"
	"github.com/daaku/go.fbapp"
	"github.com/daaku/rell/redis"
	"log"
)

const (
	yes = byte('1')
	no  = byte('0')
)

var (
	app      = fbapp.Flag("empcheck")
	fields   = fbapi.Fields([]string{"is_employee"})
	yesSlice = []byte{yes}
	noSlice  = []byte{no}
)

type user struct {
	IsEmployee bool `json:"is_employee"`
}

// Check if the user is a Facebook Employee. This only available by
// special permission granted to an application by Facebook.
func IsEmployee(id uint64) bool {
	var err error
	key := fmt.Sprintf("is_employee:%d", id)
	item, err := redis.Client().Get(key)
	if err != nil {
		log.Printf("Error in redis.Get error: %s", err)
	} else if item != nil && len(item) > 0 {
		if item[0] == yes {
			return true
		} else if item[0] == no {
			return false
		} else {
			log.Printf("Unknown redis key/value for EmpCheck: %s=%s", key, item)
		}
	}

	user := &user{}
	err = fbapi.Get(user, fmt.Sprintf("/%d", id), app, fields)
	if err != nil {
		log.Printf("Ignoring error in IsEmployee check: %s", err)
		return false
	}
	value := noSlice
	if user.IsEmployee {
		value = yesSlice
	}
	err = redis.Client().Set(key, value)
	if err != nil {
		log.Printf("Error in cache.Set: %s", err)
	}
	return user.IsEmployee
}
