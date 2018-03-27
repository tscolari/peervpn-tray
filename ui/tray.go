package ui

import (
	"github.com/getlantern/systray"
	"github.com/tscolari/peervpn-tray/peervpn"

	log "github.com/sirupsen/logrus"
)

type PeervpnRunner interface {
	Start() (chan peervpn.State, error)
	Stop() error
}

func New(runner PeervpnRunner) *Tray {
	return &Tray{
		runner:   runner,
		controls: &controls{},
	}
}

type Tray struct {
	runner    PeervpnRunner
	controls  *controls
	stateChan chan peervpn.State
}

type controls struct {
	labelConnected    *systray.MenuItem
	labelConnecting   *systray.MenuItem
	labelDisconnected *systray.MenuItem

	menuStart *systray.MenuItem
	menuStop  *systray.MenuItem
	menuQuit  *systray.MenuItem
}

func (t *Tray) OnReady() {
	systray.SetTitle("PVPN")

	t.controls.labelConnected = systray.AddMenuItem("Connected", "Connected")
	t.controls.labelConnecting = systray.AddMenuItem("Connecting", "Connecting")
	t.controls.labelDisconnected = systray.AddMenuItem("Disconnected", "Disconnected")
	t.controls.menuStart = systray.AddMenuItem("Start", "Start PeerVPN")
	t.controls.menuStop = systray.AddMenuItem("Stop", "Stop PeerVPN")
	t.controls.menuQuit = systray.AddMenuItem("Quit", "Exit from application")
	systray.AddSeparator()
	t.controls.labelConnected.Disable()
	t.controls.labelConnecting.Disable()
	systray.AddSeparator()
	t.controls.labelDisconnected.Disable()

	t.transitionToDisconnected()

	go t.watchForStart()
	go t.watchForStop()
	go t.watchForQuit()
}

func (t *Tray) OnExit() {
	_ = t.runner.Stop()
}

func (t *Tray) watchForStart() {
	for {
		<-t.controls.menuStart.ClickedCh

		var err error
		t.stateChan, err = t.runner.Start()
		if err != nil {
			log.Error("Failed to start peervpn: ", err)
			continue
		}

		t.transitionToConnecting()
		go func() {
			for newState := range t.stateChan {
				if newState.Connected {
					t.transitionToConnected()
				} else {
					t.transitionToConnecting()
				}
			}
		}()
	}
}

func (t *Tray) watchForStop() {
	for {
		<-t.controls.menuStop.ClickedCh

		close(t.stateChan)
		if err := t.runner.Stop(); err != nil {
			log.Error("Failed to stop peervpn: ", err)
		}
		t.transitionToDisconnected()
	}
}

func (t *Tray) watchForQuit() {
	for {
		<-t.controls.menuQuit.ClickedCh
		systray.Quit()
	}
}

func (t *Tray) transitionToConnected() {
	systray.SetTitle("PVPN C")
	t.controls.labelConnecting.Hide()
	t.controls.labelDisconnected.Hide()
	t.controls.labelConnected.Show()

	t.controls.menuStart.Disable()
	t.controls.menuStop.Enable()
}

func (t *Tray) transitionToConnecting() {
	systray.SetTitle("PVPN .")

	t.controls.labelDisconnected.Hide()
	t.controls.labelConnected.Hide()
	t.controls.labelConnecting.Show()

	t.controls.menuStart.Disable()
	t.controls.menuStop.Enable()
}

func (t *Tray) transitionToDisconnected() {
	systray.SetTitle("PVPN D")
	t.controls.labelConnecting.Hide()
	t.controls.labelConnected.Hide()
	t.controls.labelDisconnected.Show()

	t.controls.menuStart.Enable()
	t.controls.menuStop.Disable()
}
