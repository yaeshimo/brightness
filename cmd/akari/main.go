// command for control brightness.
//
//	akari -help
//
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/yaeshimo/brightness"
)

const (
	Name    = "akari"
	Version = "0.0.2"
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
		c: `Increment brightness 10%`,
		e: Name + " -inc",
	},
	{
		c: "Same results with -list",
		e: Name,
	},
}

func makeUsage(w *io.Writer) func() {
	return func() {
		flag.CommandLine.SetOutput(*w)
		fmt.Fprintf(*w, "Usage:\n")
		fmt.Fprintf(*w, "  %s [Options]\n", Name)
		fmt.Fprintf(*w, "  %s -set [NUMBER|max|mid|min]\n", Name)
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

	list  bool
	index int

	get    bool
	getmax bool

	set string
	inc bool
	dec bool
}

func init() {
	flag.BoolVar(&opt.help, "help", false, "Display this message")
	flag.BoolVar(&opt.version, "version", false, "Display version")

	flag.BoolVar(&opt.list, "list", false, "List candidate devices")
	flag.IntVar(&opt.index, "index", 0, "Specify device index")

	flag.BoolVar(&opt.get, "get", false, "Value of current brightness")
	flag.BoolVar(&opt.getmax, "getmax", false, "Value of max brightness")

	flag.StringVar(&opt.set, "set", "", "Set brightness [NUMBER|min|mid|max]")
	flag.BoolVar(&opt.inc, "inc", false, `Increment brightness 10%`)
	flag.BoolVar(&opt.dec, "dec", false, `Decrement brightness 10%`)
}

func state(devices ...*brightness.Device) (string, error) {
	if len(devices) == 0 {
		var err error
		devices, err = brightness.ReadDeviceAll()
		if err != nil {
			return "", err
		}
	}
	var str string
	for i, device := range devices {
		str += fmt.Sprintf("Index: %d\n", i)
		str += fmt.Sprintf("\tName: %q\n", device.Name())
		current, err := device.Current()
		if err != nil {
			return "", err
		}
		str += fmt.Sprintf("\tCurrent: %d\n", current)
		str += fmt.Sprintf("\tMax: %d\n", device.Max())
		str += fmt.Sprintf("\tMid: %d\n", device.Mid())
		str += fmt.Sprintf("\tMin: %d\n", device.Min())
	}
	return str, nil
}

func printState(devices ...*brightness.Device) error {
	str, err := state(devices...)
	if err != nil {
		return err
	}
	_, err = fmt.Print(str)
	return err
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
	case opt.list || flag.NFlag() == 0:
		return printState()
	}

	device, err := brightness.ReadDeviceIndex(opt.index)
	if err != nil {
		return err
	}

	switch {
	case opt.get:
		i, err := device.Current()
		if err != nil {
			return err
		}
		_, err = fmt.Println(i)
		return err
	case opt.getmax:
		_, err := fmt.Println(device.Max())
		return err
	}

	switch {
	case opt.set != "":
		switch opt.set {
		case "max":
			return device.SetMax()
		case "mid":
			return device.SetMid()
		case "min":
			return device.SetMin()
		default:
			i, err := strconv.Atoi(opt.set)
			if err != nil {
				return err
			}
			if i < 0 {
				return errors.New("can not set negative number " + opt.set)
			}
			return device.Set(uint(i), false)
		}
	case opt.inc:
		return device.Inc10Percent()
	case opt.dec:
		return device.Dec10Percent()
	default:
		return errors.New("arguments not enough")
	}
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
