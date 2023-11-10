# Curveadm CLI 开发

## 命令行功能项目结构

cli 目录中存放curveadm 对象的定义及相关方法，供各类开发的命令使用。

command 目录中存放各类命令(及其子命令)的实现。

curveadm.go 为`curveadm` 命令的执行入口同时执行相关审计工作。
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

更多`Command`字段及用法详见[Cobra](https://github.com/spf13/cobra)库。

### flag使用

Cobra支持自定义参数的解析：
```
cmd.PersistentFlags().BoolP("help", "h", false, "Print usage")
cmd.Flags().BoolVarP(&options.debug, "debug", "d", false, "Print debug information")
```
PersistentFlags()用于全局flag，可以被当前命令及其子命令使用。

Flags()则用于定义本地flag，仅用于当前命令。

### hook函数

cobra的Command对象支持自定义hook函数（PreRun和PostRun字段），在`run`命令执行前后运行hook函数。如下所示：
```
cmd := &cobra.Command{
  Use:   "enter ID",
  Short: "Enter service container",
  Args:  utils.ExactArgs(1),
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
hook函数会在`enter`命令执行前校验ID参数是否正确。

### 子命令
cobra允许在当前命令下添加子命令，通过当前命令对象的`AddCommand`函数。如下所示：
```
cmd.AddCommand(
  client.NewClientCommand(curveadm),         // curveadm client ...
  cluster.NewClusterCommand(curveadm),       // curveadm cluster ...
  config.NewConfigCommand(curveadm),         // curveadm config ...

  NewAuditCommand(curveadm),      // curveadm audit
  NewCleanCommand(curveadm),      // curveadm clean
  NewCompletionCommand(curveadm), // curveadm completion

  // commonly used shorthands
  hosts.NewSSHCommand(curveadm),      // curveadm ssh
  hosts.NewPlaybookCommand(curveadm), // curveadm playbook
  client.NewMapCommand(curveadm),     // curveadm map
)
```
可以将多个命令添加至一个主命令，再将主命令添加至根命令，例如`curveadm client ...`；也可以直接将命令添加至根命令，例如`curveadm audit`。

### curveadm utils
curveadm 已定义和开发一些utils供命令开发使用，在[curveadm/internal/utils目录下](https://github.com/opencurve/curveadm/tree/develop/internal/utils)。如下所示：
```
cmd := &cobra.Command{
  Use:   "client",
  Short: "Manage client",
  Args:  cliutil.NoArgs,
  RunE:  cliutil.ShowHelp(curveadm.Err()),
}
```
`cliutil.NoArgs`指明该命令不包含任何参数（子命令除外）；`cliutil.ShowHelp`函数在该命令运行时展示定义的帮助选项。
更多util及用法请参考[utils](https://github.com/opencurve/curveadm/tree/develop/internal/utils)。