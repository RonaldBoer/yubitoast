package main

import (
	"flag"
	"fmt"
	"os/exec"
	"regexp"
	"runtime"

	"github.com/pkg/errors"

	"github.com/hpcloud/tail"
	"github.com/martinlindhe/notify"
)

var pkAuthRegexp = regexp.MustCompile("PKAUTH OPENPGP\\.3$")
var pkSignRegexp = regexp.MustCompile("PKSIGN --hash=.+ OPENPGP\\.1$")

var fLogfile = flag.String("logfile", "/var/log/gpg-agent.log", "path to gpg-agent.log")
var fVerbose = flag.Bool("verbose", false, "verbose logging")
var fPopup = flag.Bool("popup", false, "show popup (if supported)")

const popupTimeout = 3 // seconds

func main() {
	flag.Parse()

	t, err := tail.TailFile(*fLogfile, tail.Config{
		Follow: true,
		// seek to end of file
		Location: &tail.SeekInfo{
			Offset: 0,
			Whence: 2,
		},
	})
	if err != nil {
		panic(errors.Wrapf(err, "tail gpg-agent log [%s]", (fLogfile)))
	}

	usePopup := false
	if *fPopup {
		switch {
		case runtime.GOOS == "darwin":
			usePopup = true
		default:
			fmt.Println("-popup not supported on", runtime.GOOS)
		}
	}

	// now we wait
	for line := range t.Lines {
		if *fVerbose {
			fmt.Println(line.Text)
		}

		if pkAuthRegexp.Match([]byte(line.Text)) {
			showNotification("Authenticate", usePopup)
		} else if pkSignRegexp.Match([]byte(line.Text)) {
			showNotification("Sign", usePopup)
		}
	}
}

func showNotification(msg string, usePopup bool) {
	appName := "Yubikey Touch Requested"

	switch {
	case runtime.GOOS == "darwin" && usePopup:
		// arg := fmt.Sprintf("tell app \"System Events\" to display dialog \"%s\" buttons \"Close\" with title \"%s\" with icon caution giving up after (%d)", appName, msg, popupTimeout)
		arg := fmt.Sprintf("tell application (path to frontmost application as text) to display dialog \"%s\" buttons {\"Close\"} with title \"%s\" with icon caution giving up after (%d)", appName, msg, popupTimeout)
		cmd := exec.Command("osascript", "-e", arg)
		err := cmd.Run()
		if err != nil {
			fmt.Println(errors.Wrapf(err, "osascript").Error())
		}
	default:
		notify.Notify("Yubikey Touch Requested", "", msg, "gnupg.png")
		notify.Notify(appName, "", msg, "gnupg.png")
	}
}
