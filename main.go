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
int
ExitApp(void) {
	[NSApp terminate:nil];
	return 0;
}
*/
import "C"

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/urfave/cli"
)

// Version is the SHA of the git commit from which this binary was built.
var Version string

func main() {
	log.SetLevel(log.ErrorLevel)
	log.SetFormatter(&log.TextFormatter{
		DisableTimestamp: true,
	})
	log.SetOutput(os.Stderr)

	app := cli.NewApp()

	app.Name = "sudolikeaboss"
	app.Version = Version
	app.Usage = "use 1password from the terminal with ease"
	app.Action = func(c *cli.Context) {
		go func() {
			run()
			C.ExitApp()
		}()
		C.StartApp()
	}

	_ = app.Run(os.Args)
}
