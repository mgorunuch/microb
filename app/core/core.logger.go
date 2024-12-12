package core

import (
	"context"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"github.com/mgorunuch/microb/app/core/zaputils"
	"golang.org/x/term"
	"os"
	"strings"
)

var loggerQuietFlag = flag.Bool("quiet", false, "Disable logging")
var Logger = zaputils.InitLogger().Sugar()

func Closer(f func() error) func() {
	return func() {
		err := f()
		if err != nil {
			Logger.Error(err)
		}
	}
}

func CtxCloser(ctx context.Context, f func(context.Context) error) func() {
	return func() {
		err := f(ctx)
		if err != nil {
			Logger.Error(err)
		}
	}
}

func FatalErr(err error) {
	if err != nil {
		Logger.Fatal(err)
	}
}

func Fatal1Err[T any](v T, err error) T {
	if err != nil {
		Logger.Fatal(err)
	}
	return v
}

func init() {
	flag.Parse()
}

func LoggerInit() {
	if *loggerQuietFlag {
		color.NoColor = true
	} else {
		PrintLogo()
	}
}

var logo = `
███╗   ███╗██╗ ██████╗██████╗  ██████╗ ██████╗ 
████╗ ████║██║██╔════╝██╔══██╗██╔═══██╗██╔══██╗
██╔████╔██║██║██║     ██████╔╝██║   ██║██████╔╝
██║╚██╔╝██║██║██║     ██╔══██╗██║   ██║██╔══██╗
██║ ╚═╝ ██║██║╚██████╗██║  ██║╚██████╔╝██████╔╝
╚═╝     ╚═╝╚═╝ ╚═════╝╚═╝  ╚═╝ ╚═════╝ ╚═════╝ 
`

var padding = 10

func PrintLogo() {
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return
	}

	var maxLogoLineLenght int

	// Center logo
	fmt.Print(strings.Repeat("\n", 1))
	lines := strings.Split(logo, "\n")
	for _, line := range lines {
		fmt.Print(strings.Repeat(" ", padding))
		fmt.Print(color.YellowString(line))
		fmt.Print("\n")

		if len(line) > maxLogoLineLenght {
			maxLogoLineLenght = len(line)
		}
	}
	fmt.Print(strings.Repeat("\n", 1))
}
