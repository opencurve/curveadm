#Curveadm CLI Development

## Command line function project structure

The definition and related methods of the curveadm object are stored in the cli directory for use by various development commands.

The command directory stores the implementation of various commands (and their subcommands).

curveadm.go is the execution entry point of the `curveadm` command and also performs related audit work.
```
cli
├── cli
│ ├── cli.go
│ └── version.go
├── command
│  ├── audit.go
│ ├── clean.go
│ ├── client
│ ├── cluster
│ ...
└── curveadm.go
```

## Cobra Library

Curveadm CLI is developed based on the [Cobra](https://github.com/spf13/cobra) library, a Go language library for creating CLI command line programs.

### Basic use of Cobra

Create a root command using Cobra (print `root commend` on the command line):
```
package main

import (
   "fmt"
   "os"
   "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
   Use: "hugo",
   Short: "Hugo is a very fast static site generator",
   Long: `A Fast and Flexible Static Site Generator built with
            love by spf13 and friends in Go.
            Complete documentation is available at https://gohugo.io/documentation/`,
   Run: func(cmd *cobra.Command, args []string) {
     fmt.Println("root commend")
   },
}

func Execute() {
   if err := rootCmd.Execute(); err != nil {
     fmt.Fprintln(os.Stderr, err)
     os.Exit(1)
   }
}

func main() {
   rootCmd.Execute()
}
```
- Use field sets the name of the command
- Short field sets a short description
- Long field setting detailed description
- The Run field sets the function when executing the command

For more details on the `Command` field and usage, please see the [Cobra](https://github.com/spf13/cobra) library

### flag usage

Cobra supports parsing of custom parameters:
```
cmd.PersistentFlags().BoolP("help", "h", false, "Print usage")
cmd.Flags().BoolVarP(&options.debug, "debug", "d", false, "Print debug information")
```
PersistentFlags() is used for global flags and can be used by the current command and its subcommands.

Flags() is used to define local flags, only for the current command.

### hook function

Cobra's Command object supports custom hook functions (PreRun and PostRun fields), and the hook function is run before and after the `run` command is executed. As follows:
```
cmd := &cobra.Command{
   Use: "enter ID",
   Short: "Enter service container",
   Args: utils.ExactArgs(1),
   PreRunE: func(cmd *cobra.Command, args []string) error {
     options.id = args[0]
     return curveadm.CheckId(options.id)
   },
   RunE: func(cmd *cobra.Command, args []string) error {
     return runEnter(curveadm, options)
   },
   DisableFlagsInUseLine: true,
}
```
The hook function will verify whether the ID parameter is correct before executing the `enter` command.

### Subcommands
cobra allows adding subcommands under the current command through the `AddCommand` function of the current command object. As follows:
```
cmd.AddCommand(
   client.NewClientCommand(curveadm), // curveadm client ...
   cluster.NewClusterCommand(curveadm), // curveadm cluster ...
   config.NewConfigCommand(curveadm), // curveadm config ...

   NewAuditCommand(curveadm), // curveadm audit
   NewCleanCommand(curveadm), // curveadm clean
   NewCompletionCommand(curveadm), // curveadm completion

   // commonly used shorthands
   hosts.NewSSHCommand(curveadm), // curveadm ssh
   hosts.NewPlaybookCommand(curveadm), // curveadm playbook
   client.NewMapCommand(curveadm), // curveadm map
)
```
You can add multiple commands to a main command, and then add the main command to the root command, such as `curveadm client...`; you can also add the command directly to the root command, such as `curveadm audit`.

### curveadm utils
curveadm has defined and developed some utils for command development, which are in the [curveadm/internal/utils directory](https://github.com/opencurve/curveadm/tree/develop/internal/utils). As follows:
```
cmd := &cobra.Command{
   Use: "client",
   Short: "Manage client",
   Args: cliutil.NoArgs,
   RunE: cliutil.ShowHelp(curveadm.Err()),
}
```
`cliutil.NoArgs` indicates that the command does not contain any arguments (except subcommands); the `cliutil.ShowHelp` function displays the defined help options when the command is run.
For more utils and usage, please refer to [utils](https://github.com/opencurve/curveadm/tree/develop/internal/utils).