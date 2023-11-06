package install

import (
	"github.com/opencurve/curveadm/cli/cli"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

func NewInstallCommand(curveadm *cli.CurveAdm) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Manage install",
		Args:  cliutil.NoArgs,
		RunE:  cliutil.ShowHelp(curveadm.Err()),
	}

	cmd.AddCommand(
		NewInstallToolCommand(curveadm),
	)
	return cmd
}
