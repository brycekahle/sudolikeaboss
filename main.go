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
	"os"

	"github.com/urfave/cli"
)

// Version is the SHA of the git commit from which this binary was built.
var Version string

func main() {
	app := cli.NewApp()

	app.Name = "sudolikeaboss"
	app.Version = Version
	app.Usage = "use 1password from the terminal with ease"
	app.Action = func(c *cli.Context) {
		go runSudolikeaboss()
		C.StartApp()
	}

	app.Run(os.Args)
}
