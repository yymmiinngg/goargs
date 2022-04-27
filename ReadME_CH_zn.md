# GOARGS

GOLang的参数处理工具

- 可视化模板，所见即所得
- 通行命令行格式，不改变输入习惯

## 示例

``` go
var argsArr = []string{"clone", "-c", "-b", "10240", "/etc/my.cnf", "/etc/my.cnf.bak1", "/etc/my.cnf.bak2", "--help"}

// 模板
template := `
    Usage: {{COMMAND}} {{OPTION}} <SRC> [DEST]...
    将文件克隆成多个副本, SRC 源文件, DEST... 目标文件列表（默认为"$SRC.bak"）
    
    ? -c, --check-space  ## 克隆前检查所须空间
    + -b, --buffer-size  ## 缓冲区大小(用于复制转移数据的临时内存空间大小)
    #                       默认值为：1024
    ?     --help         ## 显示帮助后退出
    ?     --version      ## 显示版本后退出
    
    更多细节及说明请访问 https://xxx.xxxxx.xx
`

// 定义变量
var SRC string
var DEST []string
var c, help, version bool
var b int

// 编译模板
args, err := Compile(template)
if err != nil {
    fmt.Println(err.Error())
    return
}

// 绑定变量
args.StringOperan("SRC", &SRC, "")
args.StringsOperan("DEST", &DEST, nil)
args.BoolOption("-c", &c, false)
args.IntOption("-b", &b, 1024)
args.BoolOption("--help", &help, false)
args.BoolOption("--version", &version, false)

// 处理参数
if err := args.Parse(argsArr, AllowUnknowOption); err != nil {
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
fmt.Printf("%12s  %v \n", "--help", help)
fmt.Printf("%12s  %v \n", "--vesrion", version)
fmt.Println("--------------------------------------------------")

if help {
    fmt.Println(args.Usage())
    return
}

if help {
    fmt.Println("v0.0.1")
    return
}
```

运行结果

``` shell
--------------------------------------------------
         SRC  /etc/my.cnf 
        DEST  [/etc/my.cnf.bak1 /etc/my.cnf.bak2]
          -c  true
          -b  10240
      --help  true
   --vesrion  false
--------------------------------------------------
Usage: clone [OPTION]... SRC [DEST]...
将文件克隆成多个副本, SRC 源文件, DEST... 目标文件列表（默认为"$SRC.bak"）

  -c, --check-space     克隆前检查所须空间
  -b, --buffer-size     缓冲区大小(用于复制转移数据的临时内存空间大小)
                        默认值为：1024
      --help            显示帮助后退出
      --version         显示版本后退出

更多细节及说明请访问 https://xxx.xxxxx.xx
```

## 使用说明

### 模板常量

| 常量 | 说明 |
|:-|:-|
| {{COMMAND}} | 控制台，将参数列表中的命令项（通常为第1项）渲染到Usage()输出中。 |
| {{OPTION}} | 参数项，将被替换成 \[OPTION\]... |

### 模板行类型

模板行分为4种类型，Usage行、选项定义行、注释行和空行

* Usage行

    以“Usage:”开头，作为程序（或命令）的使用说明，例如 `Usage: {{COMMAND}} {{OPTION}} <SRC> [DEST]...`

* 选项定义行

    定义程序的运行选项，有简称(-x)和全称(--xx-xxx)的区分，例如 `-b, --buffer-size` 前者为简称，后者为全称。
    
    ``` shell
    # 使用时，简称以空格分隔值，全称以等号(=)分隔值。
    cmd -x vv -xx-xxx=vv
    ```
    
    在选项定义后面以'##'分隔选项定义与选项注释，如：
    
    ```
    ? -c, --check-space  ## 克隆前检查所须空间
    ```

* 注释行
    
    整行注释 以'#' 开头，后面的内容都将作为注释输出

### 参数定义

参数定义必须放置在"Usage:"行。

| 格式 | 说明 |
|:-:|:-|
| <参数名> | 必要参数 |
| \[参数名\] | 可选参数，可选参数必须放在必要参数后面 |
| ... | 列表参数，可获取 `[]string` 类型的参数值，只能定义在所有参数的最后一个 |

### 选项定义

| 定义符 | 说明 |
|:-:|:-|
| \* | 必要选项 |
| \+ | 可选选项 |
| \? | 开关选项，后面不接参数值，输出为 `bool` 类型 |
