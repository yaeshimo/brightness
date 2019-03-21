// command for control brightness.
//
//	brightness -help
//
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/yaeshimo/brightness"
)

const (
	Name    = "brightness"
	Version = "0.0.1"
)

// example and comment for print usage
type examples []struct {
	c string
	e string
}

func (es *examples) Sprint() string {
	var s string
	for _, e := range *es {
		s += fmt.Sprintf("  %s\n", e.c)
		s += fmt.Sprintf("  $ %s\n\n", e.e)
	}
	return s
}

// Name string for specify command name
func makeUsage(w *io.Writer) func() {
	var eg = &examples{
		{
			c: "Display value of current brightness",
			e: Name + " -get",
		},
		{
			c: "Set brightness to max",
			e: Name + " -set max",
		},
		{
			c: "Increment brightness 10%",
			e: Name + " -inc",
		},
	}
	return func() {
		flag.CommandLine.SetOutput(*w)
		fmt.Fprintf(*w, "Usage:\n")
		fmt.Fprintf(*w, "  %s [Options]\n", Name)
		fmt.Fprintf(*w, "  %s -set [max|mid|min]\n", Name)
		fmt.Fprintf(*w, "\n")
		fmt.Fprintf(*w, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(*w, "\n")
		fmt.Fprintf(*w, "Examples:\n%s", eg.Sprint())
	}
}

var opt struct {
	help    bool
	version bool

	get    bool
	getmax bool

	set string
	inc bool
	dec bool
}

func init() {
	flag.BoolVar(&opt.help, "help", false, "Display this message")
	flag.BoolVar(&opt.version, "version", false, "Display version")

	flag.BoolVar(&opt.get, "get", false, "Value of current brightness")
	flag.BoolVar(&opt.getmax, "getmax", false, "Value of max brightness")

	flag.StringVar(&opt.set, "set", "", "Set brightness")
	flag.BoolVar(&opt.inc, "inc", false, "Increment brightness 10%")
	flag.BoolVar(&opt.dec, "dec", false, "Decrement brightness 10%")
}

func run() error {
	var usageWriter io.Writer = os.Stderr
	usage := makeUsage(&usageWriter)
	flag.Usage = usage

	flag.Parse()
	if flag.NArg() != 0 {
		flag.Usage()
		return fmt.Errorf("invalid arguments: %v", flag.Args())
	}

	switch {
	case opt.help:
		usageWriter = os.Stdout
		flag.Usage()
		return nil
	case opt.version:
		_, err := fmt.Printf("%s %s\n", Name, Version)
		return err
	}

	switch flag.NFlag() {
	case 0:
		// TODO: imple State()
		return errors.New("TODO: print state of devices and brightness")
	case 1:
		wrap := func(f func() (uint, error)) error {
			i, err := f()
			if err != nil {
				return err
			}
			_, err = fmt.Printf("%d\n", i)
			return err
		}
		switch {
		case opt.get:
			return wrap(brightness.Current)
		case opt.getmax:
			return wrap(brightness.Max)
		}

		switch {
		case opt.set != "":
			switch opt.set {
			case "max":
				return brightness.SetMax()
			case "mid":
				return brightness.SetMid()
			case "min":
				return brightness.SetMin()
			default:
				return fmt.Errorf("invalid arguments %q", opt.set)
			}
		case opt.inc:
			return brightness.Inc10Percent()
		case opt.dec:
			return brightness.Dec10Percent()
		default:
			return errors.New("unreachable")
		}
	default:
		return fmt.Errorf("too many specified flags")
	}
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
