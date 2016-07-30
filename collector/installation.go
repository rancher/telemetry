package collector

import (
	"github.com/satori/go.uuid"
	"io/ioutil"
)

const FILENAME = ".telemetry_id"

type Installation struct {
	Id string `json:"id"`
}

func GetInstallation() Installation {
	c := Installation{}
	c.getId()
	return c
}

func (c *Installation) getId() {
	var id string

	data, err := ioutil.ReadFile(FILENAME)
	if err == nil {
		id = string(data)
	} else {
		id = uuid.NewV4().String()
		ioutil.WriteFile(FILENAME, []byte(id), 0644)
	}

	c.Id = id
}
