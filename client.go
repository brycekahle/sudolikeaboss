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
	defer client.Close()

	helloResponse, err := client.Hello()
	if err != nil {
		log.Fatal(err)
	}

	if helloResponse.Action == onepass.ResponseAuthNew {
		fmt.Fprintf(os.Stderr, "The 1password helper will request registration of code: %s\n", helloResponse.Payload.Code)
		fmt.Fprintf(os.Stderr, "To complete registration. You must accept that code from the helper.\n")
		err = client.Register()
		if err != nil {
			log.Fatal(err)
		}
	}

	encClient, err := client.Authenticate()
	if err != nil {
		log.Fatal(err)
	}
	defer encClient.Close()

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

// Run the main sudolikeaboss entry point
func run() {
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
	dur := time.Duration(conf.TimeoutSecs) * time.Second
	select {
	case <-done:
		// Do nothing no need
	case <-time.After(dur):
		close(done)
		log.Fatalf("Timed out after %s", dur.String())
	}
}
