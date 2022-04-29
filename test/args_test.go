package test

import (
	"fmt"
	"testing"

	"github.com/yymmiinngg/goargs"
)

func Test(t *testing.T) {

	// 这里可以替换成 `os.Args` 以处理控制台命令行
	var argsArr = []string{"clone", "-c", "-b", "10240", "/etc/my.cnf", "/etc/my.cnf.bak1", "/etc/my.cnf.bak2", "--help"}

	// 模板
	template := `
		Usage: {{COMMAND}} {{OPTION}} <SRC> [DEST]...
		将文件克隆成多个副本, SRC 源文件(必填项), [DEST]... 目标文件列表(默认为"$SRC.bak")
		
		? -c, --check-space  # 克隆前检查所须空间
		+ -b, --buffer-size  # 缓冲区大小(用于复制转移数据的临时内存空间大小)
		#                      默认值为: 1024
		# -H, --help           显示帮助后退出
		#     --version        显示版本后退出
		
		更多细节及说明请访问 https://xxx.xxxxx.xx
	`

	// 定义变量
	var SRC string
	var DEST []string
	var c bool
	var b int

	// 编译模板
	args, err := goargs.Compile(template)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// 显示帮助
	if goargs.HasArgs(argsArr, "-H", "--help") {
		fmt.Println(args.Usage())
	}

	// 显示版本
	if goargs.HasArgs(argsArr, "--version") {
		fmt.Println("v0.0.1")
	}

	// 绑定变量
	args.StringOperan("SRC", &SRC, "")
	args.StringsOperan("DEST", &DEST, nil)
	args.BoolOption("-c", &c, false)
	args.IntOption("-b", &b, 1024)

	// 处理参数
	if err := args.Parse(argsArr, goargs.AllowUnknowOption); err != nil {
		fmt.Println(err.Error())
		fmt.Println(args.Usage())
		return
	}

	//  输出
	fmt.Println("--------------------------------------------------")
	fmt.Printf("%12s  %v \n", "SRC", SRC)
	fmt.Printf("%12s  %v \n", "DEST", DEST)
	fmt.Printf("%12s  %v \n", "-c", c)
	fmt.Printf("%12s  %v \n", "-b", b)
	fmt.Println("--------------------------------------------------")

}
