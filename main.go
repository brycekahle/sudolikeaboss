package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>
int
StartApp(void) {
	[NSAutoreleasePool new];
	[NSApplication sharedApplication];
	[NSApp setActivationPolicy:NSApplicationActivationPolicyProhibited];
	[NSApp run];
	return 0;
}
*/
import "C"

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"io/ioutil"

	"github.com/urfave/cli"
)

// Version is the SHA of the git commit from which this binary was built.
var Version string

func main() {
	log.SetOutput(ioutil.Discard)

	app := cli.NewApp()

	app.Name = "sudolikeaboss"
	app.Version = Version
	app.Usage = "use 1password from the terminal with ease"
	app.Action = func(c *cli.Context) {
		go runSudolikeaboss()
		C.StartApp()
	}

	app.Commands = []cli.Command{
		{
			Name:    "register",
			Aliases: []string{"a"},
			Usage:   "registers sudolikeaboss for subsequent uses",
			Action: func(c *cli.Context) {
				fmt.Println("Authenticating sudolikeaboss...")
				fmt.Println("")
				go runSudolikeabossRegistration()
				C.StartApp()
			},
		},
	}

	app.Run(os.Args)
}
