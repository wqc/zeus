package flag

import (
	"github.com/ogier/pflag"
)

var (
	ConfigPath string
	Help       bool
)

func Parse() {
	pflag.StringVarP(&ConfigPath, "config", "c", "conf/eye.yaml", "set zeus'eye config path")
	pflag.BoolVarP(&Help, "help", "h", false, "show zeus'eye useage")
	pflag.Parse()

}

func Useage() {
	pflag.Usage()
}
