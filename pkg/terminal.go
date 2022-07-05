package k8status

import "github.com/fatih/color"

func green(format string, a ...interface{}) string {
	return color.New(color.FgGreen).Sprintf(format, a...)
}

func red(format string, a ...interface{}) string {
	return color.New(color.FgRed).Sprintf(format, a...)
}
