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
	LogLevel       string `split_words:"true" default:"error"`

	Websocket struct {
		URI    string `default:"ws://127.0.0.1:6263/4"`
		Origin string `default:"resource://onepassword-at-agilebits-dot-com"`
	}
}

func LoadConfiguration() *Configuration {
	conf := Configuration{}
	err := envconfig.Process("SUDOLIKEABOSS", &conf)
	if err != nil {
		log.Fatal(err)
	}

	if lvl, err := log.ParseLevel(conf.LogLevel); err == nil {
		log.SetLevel(lvl)
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
		log.Fatal(err)
	}

	encClient, err := client.Authenticate(false)
	if err != nil {
		log.Fatal(err)
	}

	response, err := encClient.ShowPopup()
	if err != nil {
		log.Fatal(err)
	}

	password, err := response.GetPassword()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(password)

	done <- true
}

func registerWithOnepassword(configuration *onepass.Configuration, done chan bool) {
	// Load configuration from a file
	client, err := onepass.NewClientWithConfig(configuration)
	if err != nil {
		log.Fatal(err)
	}

	_, err = client.Authenticate(true)
	if err != nil {
		log.Fatal(err)
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
		WebsocketURI:    conf.Websocket.URI,
		WebsocketOrigin: conf.Websocket.Origin,
		StateDirectory:  conf.StateDirectory,
		DefaultHost:     conf.DefaultHost,
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
		WebsocketURI:    conf.Websocket.URI,
		WebsocketOrigin: conf.Websocket.Origin,
		StateDirectory:  conf.StateDirectory,
		DefaultHost:     conf.DefaultHost,
	}

	go registerWithOnepassword(&oc, done)

	// Close the app neatly
	<-done
	os.Exit(0)
}
