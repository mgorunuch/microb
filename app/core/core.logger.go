package core

import (
	"flag"
	"fmt"
	"github.com/fatih/color"
	"golang.org/x/term"
	"os"
	"strings"
)

var loggerQuietFlag = flag.Bool("quiet", false, "Disable logging")

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
