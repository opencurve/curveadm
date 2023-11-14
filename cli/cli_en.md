#Curveadm CLI Development

## Cobra library

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

For more details on `Command` object fields and usage, see [Cobra library--command.go](https://github.com/spf13/cobra/blob/main/command.go).

### flag usage

Cobra supports parsing of custom parameters:
```
cmd.PersistentFlags().BoolP("help", "h", false, "Print usage")
cmd.Flags().BoolVarP(&options.debug, "debug", "d", false, "Print debug information")
```
PersistentFlags() is used for global flags and can be used by the current command and its subcommands.

Flags() is used to define local flags, only for the current command.

For more details on the `flag` function and usage, see [Cobra library--command.go](https://github.com/spf13/cobra/blob/main/command.go) and [pflag library](https:// github.com/spf13/pflag).

### hook function

Cobra's Command object supports custom hook functions (PreRun and PostRun fields), and the hook function is run before and after the `run` command is executed. As follows:
```
cmd := &cobra.Command{
   Use: "root [sub]",
   Short: "My root command",
   PersistentPreRun: func(cmd *cobra.Command, args []string) {
       fmt.Printf("Inside rootCmd PersistentPreRun with args: %v\n", args)
     },
     PreRun: func(cmd *cobra.Command, args []string) {
       fmt.Printf("Inside rootCmd PreRun with args: %v\n", args)
     },
     Run: func(cmd *cobra.Command, args []string) {
       fmt.Printf("Inside rootCmd Run with args: %v\n", args)
     },
     PostRun: func(cmd *cobra.Command, args []string) {
       fmt.Printf("Inside rootCmd PostRun with args: %v\n", args)
     },
     PersistentPostRun: func(cmd *cobra.Command, args []string) {
       fmt.Printf("Inside rootCmd PersistentPostRun with args: %v\n", args)
     },
}
```
The hook function will be executed in the order of (`PersistentPreRun`, `PreRun`, `Run`, `PostRun`, `PersistentPostRun`). Note: If the subcommand does not set `Persistent*Run`, it will automatically inherit the function definition of the parent command.

### Subcommands
cobra allows nested commands, through the `AddCommand` function of the current command object. As follows:
```
rootCmd.AddCommand(versionCmd)
```
The recommended hierarchical command nesting structure is as follows:
```
├── cmd
│ ├── root.go
│ └── sub1
│ ├── sub1.go
│ └── sub2
│ ├── leafA.go
│ ├── leafB.go
│ └── sub2.go
└── main.go
```
Add the commands defined in leafA.go and leafB.go to the sub2 command.

Add the commands defined in sub2.go to the sub2 command.

Add the commands defined in sub2.go to the sub1 command.


## curveadm cli project structure
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
The `cli.go` in the cli folder defines the `curveadm` object and related methods, which run through all curveadm cli command development.
```
type CurveAdm struct {
   rootDir string
   dataDir string
   ...
}
func NewCurveAdm() (*CurveAdm, error) {
   curveadm := &CurveAdm{
     rootDir: rootDir,
     ...
   }
   ...
   return curveadm, nil
}
```

The command directory stores command implementations at each level.
```
├── audit.go
├── client
│ ├── cmd.go
│ ├── enter.go
│ └── unmap.go
├── cluster
│ ├── add.go
│ ├── cmd.go
├── cmd.go
├── deploy.go
```
In curveadm cli, the root command of each layer is defined in `cmd.go`. The root command is only responsible for registering subcommands and providing help information, and does not participate in actual work operations.
```
cli\command\cmd.go

func NewCurveAdmCommand(curveadm *cli.CurveAdm) *cobra.Command {
   ...
   cmd := &cobra.Command{
     Use: "curveadm [OPTIONS] COMMAND [ARGS...]",
     RunE: func(cmd *cobra.Command, args []string) error {
       if options.debug {
          return errno.List()
       } else if options.upgrade {
return tools.Upgrade2Latest(cli.Version)
       } else if len(args) == 0 {
return cliutil.ShowHelp(curveadm.Err())(cmd, args)
       }
       return fmt.Errorf("curveadm: '%s' is not a curveadm command.\n"+"See 'curveadm --help'", args[0])
     },
    ...
   }
   ...
   addSubCommands(cmd, curveadm)
   return cmd
}
################################################ ##############
cli\command\client\cmd.go

func NewClientCommand(curveadm *cli.CurveAdm) *cobra.Command {
   cmd := &cobra.Command{
       Use: "client",
       ...
       RunE: cliutil.ShowHelp(curveadm.Err()),
   }
   cmd.AddCommand(
       NewMapCommand(curveadm),
       ...
   )
   return cmd
}

```

curveadm.go defines the execution function of the `curveadm` root command and performs related audit work.
```
func Execute() {
   curveadm, err := cli.NewCurveAdm()
   ...
   id := curveadm.PreAudit(time.Now(), os.Args[1:])
   cmd := command.NewCurveAdmCommand(curveadm)
   err = cmd.Execute()
   curveadm.PostAudit(id, err)
   if err != nil {
     os.Exit(1)
   }
}
```

The entrance to the `curveadm` main program is under the [curveadm folder](https://github.com/opencurve/curveadm/tree/develop/cmd/curveadm). You can execute the operation and execution of `curveadm` in this directory. compile
```
func main() {
   cli.Execute()
}
```

### curveadm general tools
curveadm has defined and developed some common commands for development use, which are in the [curveadm/internal directory](https://github.com/opencurve/curveadm/tree/develop/internal). As follows:
```
import (
   cliutil "github.com/opencurve/curveadm/internal/utils"
   ...
)

cmd := &cobra.Command{
   Use: "client",
   Short: "Manage client",
   Args: cliutil.NoArgs,
   RunE: cliutil.ShowHelp(curveadm.Err()),
}
```
`cliutil.NoArgs` indicates that the command does not contain any arguments (except subcommands); the `cliutil.ShowHelp` function displays the defined help options when the command is run.

For more common commands and usage, please refer to [internal folder](https://github.com/opencurve/curveadm/tree/develop/internal).