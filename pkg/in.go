package resource

import (
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/tylerrasor/defectdojo-resource/client"
)

func (c *Concourse) Get() error {
	logrus.SetOutput(c.stderr)

	request, err := DecodeToGetRequest(c)
	if err != nil {
		return fmt.Errorf("invalid payload: %s", err)
	}

	if request.Source.Debug {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Debugln("debug logging on")
	}

	logrus.Debugln("creating http client")
	client, err := client.NewDefectdojoClient(request.Source.DefectDojoUrl, request.Source.Username, request.Source.Password, request.Source.ApiKey)
	if err != nil {
		return fmt.Errorf("error creating client to interact with defectdojo: %s", err)
	}

	something, err := client.GetSomethingForIn()
	if err != nil {
		return fmt.Errorf("error getting something: %s", err)
	}
	logrus.Debugln(something)

	if err := OutputVersionToConcourse(c); err != nil {
		return fmt.Errorf("error encoding response to JSON: %s", err)
	}

	return nil
}

func DecodeToGetRequest(c *Concourse) (*GetRequest, error) {
	decoder := json.NewDecoder(c.stdin)
	decoder.DisallowUnknownFields()

	var req GetRequest
	if err := decoder.Decode(&req); err != nil {
		return nil, err
	}
	return &req, nil
}
