package cli

import (
	"bytes"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

var _ = Describe("SetupCommandSilence", func() {
	It("returns nil error and logs a success message", func() {
		cmd := &cobra.Command{
			Use: "test",
			RunE: func(cmd *cobra.Command, args []string) error {
				return nil
			},
		}
		// Reset the command arguments to prevent passing ginkgo cli flags.
		// By default ginkgo adds --test.timeout flag which interferes with the command execution.
		cmd.SetArgs([]string{})

		SetupCommandSilence(cmd)
		err := cmd.Execute()

		Expect(err).To(BeNil())
	})

	It("returns a usage error", func() {
		var stdout, stderr bytes.Buffer

		cmd := &cobra.Command{
			Use: "test",
		}
		cmd.SetArgs([]string{"--unknown"})
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)

		SetupCommandSilence(cmd)
		err := cmd.Execute()

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("error setting command flags: unknown flag: --unknown"))
		Expect(err.Error()).To(ContainSubstring(cmd.UsageString()))
		Expect(stdout.String()).To(BeEmpty())
		Expect(stderr.String()).To(BeEmpty())
	})

	It("returns the unexpected error", func() {
		var stdout, stderr bytes.Buffer

		cmd := &cobra.Command{
			Use: "test",
			RunE: func(cmd *cobra.Command, args []string) error {
				return errors.New("unexpected error")
			},
		}
		// Reset the command arguments to prevent passing ginkgo cli flags.
		// By default ginkgo adds --test.timeout flag which interferes with the command execution.
		cmd.SetArgs([]string{})
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)

		SetupCommandSilence(cmd)
		err := cmd.Execute()

		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError("unexpected error"))
		Expect(err.Error()).NotTo(ContainSubstring("wrong command usage"))
		Expect(err.Error()).NotTo(ContainSubstring(cmd.UsageString()))
		Expect(stdout.String()).To(BeEmpty())
		Expect(stderr.String()).To(BeEmpty())
	})
})
