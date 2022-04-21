package goargs

import (
	"fmt"
	"testing"
)

func Test(t *testing.T) {
	template := `
		Usage: {{COMMAND}} {{OPTION}}... <URL> [FILE1] [FILE2]
		Concatenate FILE(s), or standard input, to standard output.

		+ -o, --output             output dir
		? -A, --show-all           equivalent to -vET
		+ -b, --number-nonblank    number nonempty output lines
		? -e                       equivalent to -vE
		? -E, --show-ends          display $ at end of each line
		+ -n, --number             number all output lines
		+ -s, --squeeze-blank      suppress repeated empty output lines
		? -t                       equivalent to -vT
		* -T, --show-tabs          display TAB characters as ^I
		? -u                       (ignored)
		? -v, --show-nonprinting   use ^ and M- notation, except for LFD and TAB
		?     --help               display this help and exit
		?     --version            output version information and exit

		With no FILE, or when FILE is -, read standard input.

		Examples:
		cat f - g  Output f's contents, then standard input, then g's contents.
		cat        Copy standard input to standard output.

		Report cat bugs to bug-coreutils@gnu.org
		GNU coreutils home page: <http://www.gnu.org/software/coreutils/>
		General help using GNU software: <http://www.gnu.org/gethelp/>
		For complete documentation, run: info coreutils 'cat invocation'
	`

	var outputDir, FILE1, FILE2, URL string
	var A bool
	var s int

	args, err := Compile(template)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	args.StringVar("-o", &outputDir, "")
	args.StringVar("FILE1", &FILE1, "./")
	args.StringVar("FILE2", &FILE2, "./")
	args.StringVar("URL", &URL, "")
	args.BoolVar("-A", &A, true)
	args.IntVar("-s", &s, -1)

	var argsArr = []string{"c:/window\\main.exe", "-T", "yes", "https://www.google.com", "d:\\", "-o", "d://", "-s", "10"}

	if err := args.Parse(argsArr); err != nil {
		fmt.Println(err.Error())
		fmt.Println(args.Usage())
		return
	}

	fmt.Println(">>>>>>>> ", "-o", outputDir)
	fmt.Println(">>>>>>>> ", "FILE1", FILE1)
	fmt.Println(">>>>>>>> ", "FILE2", FILE2)
	fmt.Println(">>>>>>>> ", "URL", URL)
	fmt.Println(">>>>>>>> ", "-A", A)
	fmt.Println(">>>>>>>> ", "-s", s)

	// fmt.Println(args.Usage())
}
