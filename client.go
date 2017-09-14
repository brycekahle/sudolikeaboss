package main

import (
	"fmt"
	"os"
	"os/user"
	"path"
	"time"

	"github.com/brycekahle/sudolikeaboss/onepass"
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
)

type Configuration struct {
	TimeoutSecs    int    `split_words:"true" default:"30"`
	DefaultHost    string `split_words:"true" default:"sudolikeaboss://local"`
	StateDirectory string `split_words:"true"`

	Websocket struct {
		URI      string `default:"ws://127.0.0.1:6263/4"`
		Protocol string
		Origin   string `default:"resource://onepassword-at-agilebits-dot-com"`
	}
}

func LoadConfiguration() *Configuration {
	conf := Configuration{}
	err := envconfig.Process("SUDOLIKEABOSS", &conf)
	if err != nil {
		log.Fatal(err)
	}

	if conf.StateDirectory == "" {
		usr, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}
		conf.StateDirectory = path.Join(usr.HomeDir, ".sudolikeaboss")
	}

	return &conf
}

func retrievePasswordFromOnepassword(configuration *onepass.Configuration, done chan bool) {
	// Load configuration from a file
	client, err := onepass.NewClientWithConfig(configuration)
	if err != nil {
		log.Fatalf("Error creating onepass client: %v", err)
	}

	_, err = client.Authenticate(false)
	if err != nil {
		client.Close()
		log.Fatalf("Error authenticating: %v", err)
	}

	response, err := client.SendShowPopupCommand()
	if err != nil {
		client.Close()
		log.Fatalf("Error sending showPopup command: %v", err)
	}

	password, err := response.GetPassword()
	if err != nil {
		client.Close()
		log.Fatalf("Error getting password from response: %v", err)
	}
	fmt.Println(password)

	client.Close()
	done <- true
}

func registerWithOnepassword(configuration *onepass.Configuration, done chan bool) {
	// Load configuration from a file
	client, err := onepass.NewClientWithConfig(configuration)
	if err != nil {
		os.Exit(1)
	}
	defer client.Close()

	_, err = client.Authenticate(true)
	if err != nil {
		os.Exit(1)
	}

	fmt.Println("")
	fmt.Println("Congrats sudolikeaboss is registered!")

	done <- true
}

// Run the main sudolikeaboss entry point
func runSudolikeaboss() {
	done := make(chan bool)

	conf := LoadConfiguration()
	oc := onepass.Configuration{
		WebsocketURI:      conf.Websocket.URI,
		WebsocketOrigin:   conf.Websocket.Origin,
		WebsocketProtocol: conf.Websocket.Protocol,
		StateDirectory:    conf.StateDirectory,
		DefaultHost:       conf.DefaultHost,
	}
	go retrievePasswordFromOnepassword(&oc, done)

	// Timeout if necessary
	select {
	case <-done:
		// Do nothing no need
	case <-time.After(time.Duration(conf.TimeoutSecs) * time.Second):
		close(done)
		os.Exit(1)
	}
	// Close the app neatly
	os.Exit(0)
}

func runSudolikeabossRegistration() {
	done := make(chan bool)

	conf := LoadConfiguration()
	oc := onepass.Configuration{
		WebsocketURI:      conf.Websocket.URI,
		WebsocketOrigin:   conf.Websocket.Origin,
		WebsocketProtocol: conf.Websocket.Protocol,
		StateDirectory:    conf.StateDirectory,
		DefaultHost:       conf.DefaultHost,
	}

	go registerWithOnepassword(&oc, done)

	// Close the app neatly
	<-done
	os.Exit(0)
}
