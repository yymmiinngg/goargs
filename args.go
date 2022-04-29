package goargs

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type GoArgs struct {
	template         string            // 模板
	command          string            // 命令
	operans          []string          // 参数名称
	operans_value    []string          // 参数值
	operans_requires []string          // 参数必填项
	operans_arr_item string            // 参数数组项
	options          []string          // 选项名称
	options_values   map[string]string // 选项值 <name,value>
	options_alias    map[string]string // 选项别名 <alias,name> AND <name,alias>
	options_requires []string          // 选项必填项
	options_switchs  []string          // 开关选项
	optionVars       map[string]*ArgVar
	operanVars       map[string]*ArgVar
	parseOptions     []ParseOption
}

type ArgVar struct {
	varType      string
	varLink      interface{}
	defaultValue interface{}
}

type ParseOption int

const AllowUnknowOption ParseOption = 1 // 允许未知参数

// 编译参数处理模板
func Compile(template string) (*GoArgs, error) {
	goargs := GoArgs{
		template:         template,
		operans:          make([]string, 0),
		operans_value:    make([]string, 0),
		operans_requires: make([]string, 0),
		operans_arr_item: "",
		options:          make([]string, 0),
		options_values:   make(map[string]string, 0),
		options_alias:    make(map[string]string, 0),
		options_requires: make([]string, 0),
		options_switchs:  make([]string, 0),
		optionVars:       make(map[string]*ArgVar, 0),
		operanVars:       make(map[string]*ArgVar, 0),
		parseOptions:     make([]ParseOption, 0),
	}
	lines := strings.Split(template, "\n")
	li := 0
	for _, line := range lines {
		li++
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 使用方法首行
		if strings.Index(line, "Usage:") == 0 {
			// 参数名称格式
			regOperanName, _ := regexp.Compile("^[a-zA-Z]+[a-zA-Z0-9_\\-]*$")

			// 普通参数
			reg, _ := regexp.Compile("(\\[[^\\[\\]]+\\]\\.\\.\\.)|(\\[[^\\[\\]]+\\])")
			operan := reg.FindAll([]byte(line), 1000)
			// 必须参数
			regRequire, _ := regexp.Compile("(<[^<>]+>\\.\\.\\.)|(<[^<>]+>)")
			operanRequire := regRequire.FindAll([]byte(line), 1000)

			// 检查顺序
			if len(operan) > 0 && len(operanRequire) > 0 {
				firstName := operan[0]
				for _, requireName := range operanRequire {
					if strings.Index(line, string(requireName)) > strings.Index(line, string(firstName)) {
						return nil, fmt.Errorf("required operan '%s' at the right of '%s'", requireName, firstName)
					}
				}
			}

			if len(operanRequire) > 0 {
				for _, name := range operanRequire {
					operanName := getSection(string(name), "<", ">")
					if strings.Index(string(name), "...") > 0 {
						if goargs.operans_arr_item != "" {
							return nil, fmt.Errorf("just only one list operan allowed in usage line")
						}
						goargs.operans_arr_item = operanName
					}
					if !regOperanName.Match([]byte(operanName)) {
						return nil, fmt.Errorf("invalid operan name '%s' in line %d", operanName, li)
					}
					goargs.operans = append(goargs.operans, operanName)
					goargs.operans_requires = append(goargs.operans_requires, operanName)
				}
			}

			if len(operan) > 0 {
				for _, name := range operan {
					operanName := getSection(string(name), "[", "]")
					if strings.Index(string(name), "...") > 0 {
						if goargs.operans_arr_item != "" {
							return nil, fmt.Errorf("just only one list operan allowed in usage line")
						}
						goargs.operans_arr_item = operanName
					}
					if !regOperanName.Match([]byte(operanName)) {
						return nil, fmt.Errorf("invalid operan name '%s' in line %d", operanName, li)
					}
					goargs.operans = append(goargs.operans, operanName)
				}
			}

			if goargs.operans_arr_item != "" && findOut(goargs.operans, goargs.operans_arr_item) != len(goargs.operans)-1 {
				return nil, fmt.Errorf("list operan '%s' need at then end of usage line", goargs.operans_arr_item)
			}
		}

		// 选项
		startChar := string(line[0])
		if strings.Index("+*?", startChar) < 0 {
			continue
		}
		option, optionAlias, err := compileOption(li, line, startChar)
		if err != nil {
			return nil, err
		}
		goargs.options = append(goargs.options, option)
		if optionAlias != "" {
			goargs.options_alias[option] = optionAlias
			goargs.options_alias[optionAlias] = option
		}
		switch startChar {
		case "+":
			break
		case "*":
			goargs.options_requires = append(goargs.options_requires, option)
			break
		case "?":
			goargs.options_switchs = append(goargs.options_switchs, option)
			break
		}
	}

	return &goargs, nil
}

// 使用方法
func (goargs *GoArgs) Usage() string {
	lines := strings.Split(strings.TrimSpace(goargs.template), "\n")
	text := ""
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			text += "\n"
			continue
		}
		if strings.Index(line, "Usage:") != -1 {
			line = strings.ReplaceAll(line, "<", "")
			line = strings.ReplaceAll(line, ">", "")
		}
		startChar := string(line[0])
		if strings.Index("+*?", startChar) != -1 {
			line = " " + line[1:]
			line = strings.Replace(line, "#", " ", 1)
		}
		if startChar == "#" {
			line = " " + line[1:]
		}

		cmd := goargs.command
		cmd1 := getRightOuter(goargs.command, "\\")
		cmd2 := getRightOuter(goargs.command, "/")
		if len(cmd1) != 0 {
			cmd = cmd1
		}
		if len(cmd2) != 0 && len(cmd2) < len(cmd) {
			cmd = cmd2
		}
		line = strings.ReplaceAll(line, "{{COMMAND}}", cmd)
		line = strings.ReplaceAll(line, "{{OPTION}}", "[OPTION]...")
		text += line + "\n"
	}
	return text
}

// 处理参数
func (goargs *GoArgs) Parse(args []string, parseOptions ...ParseOption) error {
	goargs.parseOptions = append(goargs.parseOptions, parseOptions...)
	li := -1
	for true {
		li++
		if len(args) <= li {
			break
		}
		item := args[li]

		if li == 0 {
			goargs.command = item
			continue
		}

		// 选项格式 -name value
		if isOptionShortName(item) {
			name := item

			// 参数是否定义
			if findOut(goargs.options, name) < 0 && goargs.optionAlias(name) == "" {
				if findOutParseOption(goargs.parseOptions, AllowUnknowOption) != -1 {
					goto doOperan
				}
				return fmt.Errorf("invalid option '%s'", name)
			}

			// 存储开关项值
			if findOut(goargs.options_switchs, name) != -1 || findOut(goargs.options_switchs, goargs.optionAlias(name)) != -1 {
				goargs.options_values[name] = "on"
				continue
			}

			// 存储短参数值
			if len(args) > li+1 {
				value := args[li+1]
				// 值为参数形式
				if isOptionShortName(value) || isOptionLongName(value) {
					return fmt.Errorf("invalid option '%s'", name)
				}
				goargs.options_values[name] = value
				li++
				continue
			} else {
				return fmt.Errorf("invalid option '%s'", name)
			}
		}

		// 选项格式 --name=value
		if isOptionLongName(item) {
			// 获取参数名
			name := getLeft(item, "=")

			if name == "" {
				name = item
				// 存储开关项值
				if findOut(goargs.options_switchs, name) != -1 || findOut(goargs.options_switchs, goargs.optionAlias(name)) != -1 {
					goargs.options_values[name] = "on"
					continue
				}
				if findOutParseOption(goargs.parseOptions, AllowUnknowOption) != -1 {
					goto doOperan
				}
				return fmt.Errorf("unrecognized option '%s'", name)
			}

			// 参数是否定义
			if findOut(goargs.options, name) < 0 && goargs.optionAlias(name) == "" {
				if findOutParseOption(goargs.parseOptions, AllowUnknowOption) != -1 {
					goto doOperan
				}
				return fmt.Errorf("invalid option '%s'", name)
			}

			// 获取参数值
			value := getRight(item, "=")
			if value == "" {
				return fmt.Errorf("unrecognized option '%s'", name)
			}

			goargs.options_values[name] = value
			continue
		}

	doOperan:
		// 参数
		goargs.operans_value = append(goargs.operans_value, item)
	}

	// 检查必须参数
	if len(goargs.operans_requires) > len(goargs.operans_value) {
		return fmt.Errorf("missing operand '%s'", goargs.operans_requires[len(goargs.operans_value)])
	}

	// 检查必须选项
	for _, name := range goargs.options_requires {
		_, ok := goargs.options_values[name]
		if !ok {
			return fmt.Errorf("missing option '%s'", name)
		}
	}

	// 自动处理参数变量
	for name, argVar := range goargs.operanVars {
		var err error
		if argVar.varType == "[]string" {
			if goargs.operans_arr_item == "" || name != goargs.operans_arr_item {
				return fmt.Errorf("list operan '%s' not found in usage", name)
			}
			err = setValue(name, argVar, goargs.OperandStrings(name, nil))
		} else {
			if name == goargs.operans_arr_item {
				return fmt.Errorf("'%s' is not a single operan in usage", name)
			}
			err = setValue(name, argVar, goargs.Operand(name, ""))
		}
		if err != nil {
			return fmt.Errorf("invalid option '%s', because %s", name, err.Error())
		}
	}

	// 自动处理选项变量
	for name, argVar := range goargs.optionVars {
		if !(isOptionLongName(name) || isOptionShortName(name)) {
			return fmt.Errorf("invalid option '%s'", name)
		}
		var err error
		value, ok := goargs.options_values[name]
		if !ok {
			name = goargs.optionAlias(name)
		}
		value, _ = goargs.options_values[name]
		err = setValue(name, argVar, value)
		if err != nil {
			return fmt.Errorf("invalid option '%s', because %s", name, err.Error())
		}
	}

	return nil
}

// 是否存在参数项
func HasArgs(args []string, options ...string) bool {
	for _, h := range options {
		if findOut(args, h) != -1 {
			return true
		}
	}
	return false
}

// 所有参数
func (it *GoArgs) AllOperand() []string {
	return append([]string{}, it.operans_value...)
}

// 所有选项
func (it *GoArgs) AllOption() map[string]string {
	cloneTags := make(map[string]string)
	for k, v := range it.options_values {
		cloneTags[k] = v
	}
	return cloneTags
}

// 字符串值
func (it *GoArgs) Option(name string, defaultValue string) string {
	value, ok := it.options_values[name]
	if ok {
		return value
	}
	value, ok = it.options_values[it.optionAlias(name)]
	if ok {
		return value
	}
	return defaultValue
}

// bool值
func (it *GoArgs) Has(name string, defaultValue bool) bool {
	if findOut(it.options_switchs, name) < 0 {
		return defaultValue
	}
	value, ok := it.options_values[name]
	return ok && (value == "on" || value == "yes" || value == "true")
}

// 根据命名获取参数值
func (it *GoArgs) Operand(name string, defaultValue string) string {
	li := findOut(it.operans, name)
	if li < 0 {
		return defaultValue
	}
	if li >= len(it.operans_value) {
		return defaultValue
	}
	if it.operans_arr_item == name {
		return defaultValue
	}
	return it.operans_value[li]
}

// 根据命名获取参数值(数组)
func (it *GoArgs) OperandStrings(name string, defaultValue []string) []string {
	li := findOut(it.operans, name)
	if li == -1 {
		return defaultValue
	}
	if it.operans_arr_item != name {
		return defaultValue
	}
	if li >= len(it.operans_value) {
		return defaultValue
	}
	return it.operans_value[li:]
}

// 根据位置获取参数值
func (it *GoArgs) OperandAt(index int, defaultValue string) string {
	if index < len(it.operans_value) {
		return it.operans_value[index]
	}
	return defaultValue
}

// 参数
func (it *GoArgs) StringOperan(argName string, strVar *string, defaultValue string) {
	it.operanVars[argName] = &ArgVar{
		varType:      "string",
		varLink:      strVar,
		defaultValue: defaultValue,
	}
}

func (it *GoArgs) StringsOperan(argName string, strVar *[]string, defaultValue []string) {
	it.operanVars[argName] = &ArgVar{
		varType:      "[]string",
		varLink:      strVar,
		defaultValue: defaultValue,
	}
}

func (it *GoArgs) IntOperan(argName string, strVar *string, defaultValue string) {
	it.operanVars[argName] = &ArgVar{
		varType:      "int",
		varLink:      strVar,
		defaultValue: defaultValue,
	}
}

func (it *GoArgs) Int32Operan(argName string, int32Var *int32, defaultValue int32) {
	it.operanVars[argName] = &ArgVar{
		varType:      "int32",
		varLink:      int32Var,
		defaultValue: defaultValue,
	}
}

func (it *GoArgs) Int64Operan(argName string, int64Var *int64, defaultValue int64) {
	it.operanVars[argName] = &ArgVar{
		varType:      "int64",
		varLink:      int64Var,
		defaultValue: defaultValue,
	}
}

func (it *GoArgs) Float32Operan(argName string, float32Var *float32, defaultValue float32) {
	it.operanVars[argName] = &ArgVar{
		varType:      "float32",
		varLink:      float32Var,
		defaultValue: defaultValue,
	}
}

func (it *GoArgs) Float64Operan(argName string, float64Var *float64, defaultValue float64) {
	it.operanVars[argName] = &ArgVar{
		varType:      "float64",
		varLink:      float64Var,
		defaultValue: defaultValue,
	}
}

// 选项
func (it *GoArgs) StringOption(argName string, strVar *string, defaultValue string) {
	it.optionVars[argName] = &ArgVar{
		varType:      "string",
		varLink:      strVar,
		defaultValue: defaultValue,
	}
}

func (it *GoArgs) BoolOption(argName string, boolVar *bool, defaultValue bool) {
	it.optionVars[argName] = &ArgVar{
		varType:      "bool",
		varLink:      boolVar,
		defaultValue: defaultValue,
	}
}

func (it *GoArgs) IntOption(argName string, intVar *int, defaultValue int) {
	it.optionVars[argName] = &ArgVar{
		varType:      "int",
		varLink:      intVar,
		defaultValue: defaultValue,
	}
}

func (it *GoArgs) Int32Option(argName string, int32Var *int32, defaultValue int32) {
	it.optionVars[argName] = &ArgVar{
		varType:      "int32",
		varLink:      int32Var,
		defaultValue: defaultValue,
	}
}

func (it *GoArgs) Int64Option(argName string, int64Var *int64, defaultValue int64) {
	it.optionVars[argName] = &ArgVar{
		varType:      "int64",
		varLink:      int64Var,
		defaultValue: defaultValue,
	}
}

func (it *GoArgs) Float32Option(argName string, float32Var *float32, defaultValue float32) {
	it.optionVars[argName] = &ArgVar{
		varType:      "float32",
		varLink:      float32Var,
		defaultValue: defaultValue,
	}
}

func (it *GoArgs) Float64Option(argName string, float64Var *float64, defaultValue float64) {
	it.optionVars[argName] = &ArgVar{
		varType:      "float64",
		varLink:      float64Var,
		defaultValue: defaultValue,
	}
}

func setValue(name string, argVar *ArgVar, value interface{}) error {
	var err error
	switch argVar.varType {
	case "string":
		if value == "" {
			*argVar.varLink.(*string) = argVar.defaultValue.(string)
		} else {
			*argVar.varLink.(*string) = value.(string)
		}
		break
	case "[]string":
		if value == nil || len(value.([]string)) == 0 {
			*argVar.varLink.(*[]string) = argVar.defaultValue.([]string)
		} else {
			*argVar.varLink.(*[]string) = value.([]string)
		}
		break
	case "bool":
		if value == "" {
			*argVar.varLink.(*bool) = argVar.defaultValue.(bool)
		} else {
			v := (value == "on" || value == "yes" || value == "true")
			*argVar.varLink.(*bool) = v
		}
		break
	case "int":
		if value == "" {
			*argVar.varLink.(*int) = argVar.defaultValue.(int)
		} else {
			var v int
			v, err = strconv.Atoi(value.(string))
			*argVar.varLink.(*int) = v
		}
		break
	case "int32":
		if value == "" {
			*argVar.varLink.(*int32) = argVar.defaultValue.(int32)
		} else {
			var v int64
			v, err = strconv.ParseInt(value.(string), 10, 32)
			*argVar.varLink.(*int32) = int32(v)
		}
		break
	case "int64":
		if value == "" {
			*argVar.varLink.(*int64) = argVar.defaultValue.(int64)
		} else {
			var v int64
			v, err = strconv.ParseInt(value.(string), 10, 64)
			*argVar.varLink.(*int64) = v
		}
		break
	case "float32":
		if value == "" {
			*argVar.varLink.(*float32) = argVar.defaultValue.(float32)
		} else {
			var v float64
			v, err = strconv.ParseFloat(value.(string), 32)
			*argVar.varLink.(*float32) = float32(v)
		}
		break
	case "float64":
		if value == "" {
			*argVar.varLink.(*float64) = argVar.defaultValue.(float64)
		} else {
			var v float64
			v, err = strconv.ParseFloat(value.(string), 64)
			*argVar.varLink.(*float64) = v
		}
		break
	}

	return err
}

func (it *GoArgs) optionAlias(name string) string {
	aliasName, ok := it.options_alias[name]
	if ok {
		return aliasName
	}
	return ""
}

func isOptionLongName(item string) bool {
	return strings.Index(item, "--") == 0
}

func isOptionShortName(item string) bool {
	return strings.Index(item, "-") == 0 && strings.Index(item, "--") != 0
}

func findOut(list []string, key string) int {
	for i := 0; i < len(list); i++ {
		if list[i] == key {
			return i
		}
	}
	return -1
}

func findOutParseOption(list []ParseOption, key ParseOption) int {
	for i := 0; i < len(list); i++ {
		if list[i] == key {
			return i
		}
	}
	return -1
}

func compileOption(li int, line string, start string) (string, string, error) {
	// 验证
	ok, _ := regexp.Match("^\\"+start+"( *\\-{1,2}[a-zA-Z]+[a-zA-Z0-9_\\-]*)(, *\\-{1,2}[a-zA-Z]+[a-zA-Z0-9_\\-]*)?( *#+.*)?$", []byte(line))
	if !ok {
		return "", "", fmt.Errorf("incorrect line at %d", li)
	}

	setstr := getSection(line, start, "#")
	if setstr == "" {
		setstr = getRight(line, start)
	}

	// option
	option := getLeft(setstr, ",")
	if option == "" {
		option = setstr
	}
	option = strings.TrimSpace(option)

	// optionAlias
	optionAlias := getRight(setstr, ",")
	optionAlias = strings.TrimSpace(optionAlias)

	return option, optionAlias, nil
}

func getSection(str string, start string, end string) string {
	si := strings.Index(str, start)
	if si < 0 || len(str) <= si+len(start) {
		return ""
	}
	tmp := str[si+len(start):]
	ei := strings.Index(tmp, end)
	if ei < 0 {
		return ""
	}
	return tmp[:ei]
}

func getRight(str string, start string) string {
	si := strings.Index(str, start)
	if si < 0 || len(str) <= si+len(start) {
		return ""
	}
	return str[si+len(start):]
}

func getRightOuter(str string, start string) string {
	si := strings.LastIndex(str, start)
	if si < 0 || len(str) <= si+len(start) {
		return ""
	}
	return str[si+len(start):]
}

func getLeft(str string, start string) string {
	si := strings.Index(str, start)
	if si < 0 {
		return ""
	}
	return str[:si]
}
