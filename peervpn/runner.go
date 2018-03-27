package peervpn

import (
	"bufio"
	"io"
	"os/exec"
	"regexp"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const notConnectedRegex = "0 peers connected"
const connectedRegex = "[1-9][0-9]* peer(s)? connected"

type State struct {
	Connected bool
}

type CommandRunner interface {
	Start(*exec.Cmd) error
	Kill(*exec.Cmd) error
}

func NewRunner(cmdRunner CommandRunner, configPath string) *Runner {
	return &Runner{
		configPath:         configPath,
		connectedRegexp:    regexp.MustCompile(connectedRegex),
		notConnectedRegexp: regexp.MustCompile(notConnectedRegex),
		cmdRunner:          cmdRunner,
	}
}

type Runner struct {
	configPath string
	cmdRunner  CommandRunner
	currentCmd *exec.Cmd
	state      chan State

	connectedRegexp    *regexp.Regexp
	notConnectedRegexp *regexp.Regexp
}

func (r *Runner) Start() (chan State, error) {
	log.Debug("Runner.Start")
	r.state = make(chan State)

	r.currentCmd = exec.Command("peervpn", r.configPath)
	output, err := r.currentCmd.StdoutPipe()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create output pipe")
	}

	//TODO: command gets killed? What then
	if err := r.cmdRunner.Start(r.currentCmd); err != nil {
		return nil, errors.Wrap(err, "failed to start peervpn")
	}

	go r.watchForOutput(output)

	return r.state, nil
}

func (r *Runner) Stop() error {
	log.Debug("Runner.Stop")
	err := r.cmdRunner.Kill(r.currentCmd)
	if err != nil {
		return errors.Wrap(err, "failed to stop")
	}

	return nil
}

func (r *Runner) watchForOutput(output io.Reader) {

	scanner := bufio.NewScanner(output)
	for scanner.Scan() {
		line := scanner.Bytes()
		if r.connectedRegexp.Match(line) {
			log.Debug("PeerVPN connected")
			r.state <- State{true}
			continue
		}

		if r.notConnectedRegexp.Match(line) {
			log.Debug("PeerVPN not connected")
			r.state <- State{false}
		}
	}
}
