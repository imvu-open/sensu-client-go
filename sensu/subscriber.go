package sensu

import (
	"bytes"
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"time"
)

type Subscriber struct {
	Subscription string
	Client       *Client
}

func (s *Subscriber) SetClient(c *Client) error {
	s.Client = c

	return nil
}

func (s *Subscriber) Start() error {
	funnel := strings.Join(
		[]string{
			s.Client.Config.Name(),
			CurrentVersion,
			strconv.Itoa(int(time.Now().Unix())),
		},
		"-",
	)

	log.Println(s.Subscription)
	log.Println(funnel)

	msgChan := make(chan []byte)
	stopChan := make(chan bool)

	go s.Client.Transport.Subscribe("#", s.Subscription, funnel, msgChan, stopChan)

	log.Printf("Subscribed to %s", s.Subscription)

	var b []byte
	log.Println(&msgChan)

	for {
		b = <-msgChan

		payload := make(map[string]interface{})
		result := make(map[string]interface{})

		log.Printf("Check received : %s", bytes.NewBuffer(b).String())
		json.Unmarshal(b, &payload)

		output := (&Check{payload}).Execute()

		result["client"] = s.Client.Config.Name()

		formattedOuput := make(map[string]interface{})

		formattedOuput["name"] = payload["name"]
		formattedOuput["issued"] = int(payload["issued"].(float64))
		formattedOuput["output"] = output.Output
		formattedOuput["duration"] = output.Duration
		formattedOuput["status"] = output.Status
		formattedOuput["executed"] = output.Executed

		result["check"] = formattedOuput

		p, err := json.Marshal(result)

		if err != nil {
			log.Printf("something goes wrong : %s", err.Error())
		}

		log.Printf("Payload sent: %s", bytes.NewBuffer(p).String())

		err = s.Client.Transport.Publish("direct", "results", "", p)
	}

	return nil
}