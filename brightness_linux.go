// +build linux

package brightness

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// expected locations
// root    : "/sys/class/backlight/"
// devices : "/sys/class/backlight/*/"
// files   : "/sys/class/backlight/*/{max_,}brightness"

// can modify for test
var root = "/sys/class/backlight/"

const (
	baseCurrent = "brightness"
	baseMax     = "max_brightness"
)

func init() {
	readDeviceAll = func() ([]*Device, error) {
		fis, err := ioutil.ReadDir(root)
		if err != nil {
			return nil, err
		}
		devices := make([]*Device, 0, len(fis))
		for _, fi := range fis {
			if fi.Mode()&os.ModeSymlink != 0 {
				fi, err = os.Stat(filepath.Join(root, fi.Name()))
				if err != nil {
					return nil, err
				}
			}
			if fi.IsDir() {
				internal := &device{
					root: filepath.Join(root, fi.Name()),
				}
				max, err := internal.Max()
				if err != nil {
					return nil, err
				}
				devices = append(devices, &Device{
					internal: internal,
					max:      max,
				})
			}
		}
		if len(devices) < 1 {
			return nil, errors.New("not found devices in " + root)
		}
		return devices, nil
	}
}

// for read {max_,}brightness
func readUint(file string) (uint, error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return 0, err
	}
	i, err := strconv.ParseUint(strings.TrimSpace(string(b)), 10, 64)
	return uint(i), err
}

// implement for the type internal
type device struct {
	// full path to target device directory
	root string
}

func (d *device) Name() string {
	return filepath.Base(d.root)
}

func (d *device) Current() (uint, error) {
	return readUint(filepath.Join(d.root, baseCurrent))
}

func (d *device) Max() (uint, error) {
	return readUint(filepath.Join(d.root, baseMax))
}

func (d *device) Set(ui uint) error {
	max, err := d.Max()
	if err != nil {
		return err
	}
	if ui > max {
		return errors.New("requested brightness over the max")
	}
	f, err := os.OpenFile(filepath.Join(d.root, baseCurrent), os.O_WRONLY|os.O_TRUNC, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(strconv.FormatUint(uint64(ui), 10))
	return err
}
