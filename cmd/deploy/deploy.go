package deploy

import (
	"github.com/spf13/cobra"

	"github.com/ydb-platform/ydbops/pkg/cli"
	"github.com/ydb-platform/ydbops/pkg/cmdutil"
)

func New(f cmdutil.Factory) *cobra.Command {
	opts := &Options{}

	cmd := cli.SetDefaultsOn(&cobra.Command{
		Use:   "deploy",
		Short: "Deploy YDB cluster",
		Long: `ydbops deploy [command]:
    Deploy YDB cluster on hosts like Ansible`,
		// RunE: cli.RequireSubcommand,
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.Run(f)
		},
	})

	opts.DefineFlags(cmd.PersistentFlags())

	// cmd.AddCommand(
	// 	complete.New(f),
	// 	create.New(f),
	// 	drop.New(f),
	// 	list.New(f),
	// 	refresh.New(f),
	// )

	return cmd
}
