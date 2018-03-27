package peervpn_test

import (
	"errors"
	"os/exec"
	"time"

	"code.cloudfoundry.org/commandrunner/fake_command_runner"

	"github.com/tscolari/peervpn-tray/peervpn"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Runner", func() {
	var (
		subject    *peervpn.Runner
		cmdRunner  *fake_command_runner.FakeCommandRunner
		configPath string
	)

	BeforeEach(func() {
		configPath = "/hello"
		cmdRunner = fake_command_runner.New()
		subject = peervpn.NewRunner(cmdRunner, configPath)
	})

	Describe("Start", func() {
		It("executes the correct command", func() {
			_, err := subject.Start()
			Expect(err).NotTo(HaveOccurred())

			executedCmds := cmdRunner.StartedCommands()
			Expect(executedCmds).To(HaveLen(1))

			Expect(len(executedCmds[0].Args)).To(Equal(2))
			Expect(executedCmds[0].Args[0]).To(Equal("peervpn"))
			Expect(executedCmds[0].Args[1]).To(Equal("/hello"))
		})

		Describe("the channel returned", func() {
			var peerVpnOutput string

			BeforeEach(func() {
				peerVpnOutput = ""
				delay := 50 * time.Millisecond
				cmdRunner.WhenStarting(fake_command_runner.CommandSpec{}, func(cmd *exec.Cmd) error {
					time.Sleep(delay)
					cmd.Stdout.Write([]byte(peerVpnOutput))
					return nil
				})
			})

			It("publishes a connected state when peers connected output is found", func() {
				peerVpnOutput = "\n100 peers connected\n\n"
				conState := peervpn.State{Connected: false}

				stateChan, err := subject.Start()
				Expect(err).NotTo(HaveOccurred())

				go func() {
					GinkgoRecover()
					conState = <-stateChan
				}()

				Expect(err).NotTo(HaveOccurred())
				Eventually(func() bool { return conState.Connected }).Should(BeTrue())
			})

			It("publishes a not connected state when no peers connected output is found", func() {
				peerVpnOutput = "\n0 peers connected\n\n"
				conState := peervpn.State{Connected: true}

				stateChan, err := subject.Start()
				Expect(err).NotTo(HaveOccurred())

				go func() {
					GinkgoRecover()
					conState = <-stateChan
				}()

				Expect(err).NotTo(HaveOccurred())
				Eventually(func() bool { return conState.Connected }).Should(BeFalse())
			})
		})

		Context("when the command fails to start", func() {
			It("returns an error", func() {
				cmdRunner.WhenStarting(fake_command_runner.CommandSpec{}, func(*exec.Cmd) error {
					return errors.New("failed")
				})

				_, err := subject.Start()
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(MatchRegexp("failed")))
			})
		})
	})

	Describe("Stop", func() {
		It("kills a command", func() {
			Expect(subject.Stop()).To(Succeed())
			Expect(cmdRunner.KilledCommands()).To(HaveLen(1))
		})
	})
})
