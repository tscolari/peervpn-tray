package main

import (
	"code.cloudfoundry.org/commandrunner/linux_command_runner"
	"github.com/getlantern/systray"
	"github.com/tscolari/peervpn-tray/peervpn"
	"github.com/tscolari/peervpn-tray/ui"
)

func main() {
	cmdRunner := linux_command_runner.New()
	runner := peervpn.NewRunner(cmdRunner)
	tray := ui.New(runner)
	systray.Run(tray.OnReady, tray.OnExit)
}
