# Curveadm CLI 开发

## Cobra库

Curveadm CLI是基于[Cobra](https://github.com/spf13/cobra)库（一个用于创建CLI命令行程序的Go语言库）开发的。

### Cobra的基本使用

使用Cobra创建一个根命令（在命令行打印`root commend`）:
```
package main

import (
  "fmt"
  "os"
  "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
  Use:   "hugo",
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
- Use字段设置命令的名字
- Short字段设置简短描述
- Long字段设置详细描述
- Run字段设置执行该命令时的函数

更多`Command`对象字段及用法详见[Cobra库--command.go](https://github.com/spf13/cobra/blob/main/command.go)。

### flag使用

Cobra支持自定义参数的解析：
```
cmd.PersistentFlags().BoolP("help", "h", false, "Print usage")
cmd.Flags().BoolVarP(&options.debug, "debug", "d", false, "Print debug information")
```
PersistentFlags()用于全局flag，可以被当前命令及其子命令使用。

Flags()则用于定义本地flag，仅用于当前命令。

更多`flag`函数及用法详见[Cobra库--command.go](https://github.com/spf13/cobra/blob/main/command.go)和[pflag库](https://github.com/spf13/pflag)。

### hook函数

cobra的Command对象支持自定义hook函数（PreRun和PostRun字段），在`run`命令执行前后运行hook函数。如下所示：
```
cmd := &cobra.Command{
  Use:   "root [sub]",
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
hook函数会按（`PersistentPreRun`、`PreRun`、`Run`、`PostRun`、`PersistentPostRun`）的顺序依次执行。注意：如果子命令没有设置`Persistent*Run`，则会自动继承父命令的函数定义。

### 子命令
cobra允许嵌套命令，通过当前命令对象的`AddCommand`函数。如下所示：
```
rootCmd.AddCommand(versionCmd)
```
推荐的层级命令嵌套结构如下：
```
├── cmd
│   ├── root.go
│   └── sub1
│       ├── sub1.go
│       └── sub2
│           ├── leafA.go
│           └── leafB.go
└── main.go
```
将 leafA.go 和 leafB.go 中定义的命令添加到 sub2 命令中。

将 sub2.go 中定义的命令添加到 sub1 命令中。

将 sub1.go 中定义的命令添加到 root 命令中。

最终的命令调用的主体结构如下：
```
root options
root sub1 options
root sub1 sub2 options
root sub1 sub2 leafA options
root sub1 sub2 leafB options
```


## curveadm cli项目结构
```
cli
├── cli
│   ├── cli.go
│   └── version.go
├── command
│   ├── audit.go
│   ├── clean.go
│   ├── client
│   ├── cluster
│   ...
└── curveadm.go
```
cli 文件夹的`cli.go`定义了`curveadm`对象及相关方法，贯穿所有curveadm cli的命令开发。
```
type CurveAdm struct {
  rootDir      string
  dataDir      string
  ...
}
func NewCurveAdm() (*CurveAdm, error) {
  curveadm := &CurveAdm{
    rootDir:      rootDir,
    ...
  }
  ...
  return curveadm, nil
}
```

command 目录中存放各层级命令实现。
```
├── audit.go
├── client
│   ├── cmd.go
│   ├── enter.go
│   └── unmap.go
├── cluster
│   ├── add.go
│   ├── cmd.go
├── cmd.go
├── deploy.go
```
在curveadm cli中，每层的根命令都在`cmd.go`定义。根命令只负责注册子命令以及提供帮助信息，并不参与实际工作操作。
```
cli\command\cmd.go

func addSubCommands(cmd *cobra.Command, curveadm *cli.CurveAdm) {
  cmd.AddCommand(
    client.NewClientCommand(curveadm),         // curveadm client
    ...
  )
}
func NewCurveAdmCommand(curveadm *cli.CurveAdm) *cobra.Command {
  ...
  cmd := &cobra.Command{
    Use:     "curveadm [OPTIONS] COMMAND [ARGS...]",
    ...
  }
  ...
  addSubCommands(cmd, curveadm)
  return cmd
}
################################################################
cli\command\client\cmd.go

func NewClientCommand(curveadm *cli.CurveAdm) *cobra.Command {
  cmd := &cobra.Command{
    Use:   "client",
    ...
  }
  cmd.AddCommand(
    NewMapCommand(curveadm),
    ...
  )
  ...
}
################################################################
cli\command\client\enter.go

func NewEnterCommand(curveadm *cli.CurveAdm) *cobra.Command {
  ...
  cmd := &cobra.Command{
    Use:   "enter ID",
    ...
  }
  ...
}
```
最终enter命令的调用结构如下：
```
curveadm client enter ID
```

curveadm.go 定义了`curveadm` 根命令的执行函数同时执行相关审计工作。
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

`curveadm` 主程序的入口则是在[curveadm文件夹下](https://github.com/opencurve/curveadm/tree/develop/cmd/curveadm)，可以在该目录下执行`curveadm`的运行及编译
```
func main() {
  cli.Execute()
}
```

### curveadm 通用工具
对于curveadm cli的命令开发，curveadm 提供了通用工具，如：

- cliutil.NoArgs：用于判断命令是否不包含参数

- cliutil.ShowHelp：用于在命令运行时展示帮助信息

在[curveadm/internal目录下](https://github.com/opencurve/curveadm/tree/develop/internal)。如下所示：
```
import (
  cliutil "github.com/opencurve/curveadm/internal/utils"
  ...
)

cmd := &cobra.Command{
  Use:   "client",
  Args:  cliutil.NoArgs,
  RunE:  cliutil.ShowHelp(curveadm.Err()),
}
```
`cliutil.NoArgs`指明`curveadm client`命令不包含任何参数（子命令除外）；`cliutil.ShowHelp`函数在直接运行`curveadm client`命令时展示定义的帮助选项。

更多通用命令及用法请参考[internal文件夹](https://github.com/opencurve/curveadm/tree/develop/internal)。