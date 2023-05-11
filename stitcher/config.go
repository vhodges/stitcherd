package stitcher

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

func ReadHostConfigFile(filename string) (c *Host, err error) {

	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// TODO Unmarshall JSON into Host struct pointer
	var host Host
	json.Unmarshal([]byte(content), &host)

	return &host, nil
}

func NewHostFromFile(file string) (*Host, error) {

	host, err := ReadHostConfigFile(file)

	if err != nil {
		log.Printf("Error reading config file '%s': %v", file, err)
		return  nil, err
	}

	host.Init()
	
	return host, nil
}
