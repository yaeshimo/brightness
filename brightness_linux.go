// +build linux

package brightness

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// expected locations
// root    : "/sys/class/backlight/"
// devices : "/sys/class/backlight/*/"
// files   : "/sys/class/backlight/*/{max_,}brightness"
const (
	defaultRoot        = "/sys/class/backlight/"
	defaultBaseCurrent = "brightness"
	defaultBaseMax     = "max_brightness"
)

// for read {max_,}brightness
func readUint(file string) (uint, error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return 0, err
	}
	i, err := strconv.ParseUint(strings.TrimSpace(string(b)), 10, 64)
	return uint(i), err
}

// device path
type device struct {
	// full path to target device
	dir string

	// basename current brightness
	baseCurrent string

	// basename max brightness
	baseMax string
}

func (d *device) current() (uint, error) {
	return readUint(filepath.Join(d.dir, d.baseCurrent))
}
func (d *device) max() (uint, error) {
	return readUint(filepath.Join(d.dir, d.baseMax))
}
func (d *device) set(ui uint) error {
	max, err := d.max()
	if err != nil {
		return err
	}
	if ui > max {
		return errors.New("requested brightness is over the max")
	}
	f, err := os.OpenFile(filepath.Join(d.dir, d.baseCurrent), os.O_WRONLY|os.O_TRUNC, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(strconv.FormatUint(uint64(ui), 10))
	return err
}

// implement for Brightness
// default target device is devices[0]
type brightness struct {
	// devices root
	// default "/sys/class/backlight"
	root string

	// store condidates
	devices []*device

	// picked target device from devices
	picked *device
}

func init() {
	internal = &brightness{root: defaultRoot}
}

// read from b.root
func (b *brightness) readDevices() error {
	b.devices = nil
	fis, err := ioutil.ReadDir(b.root)
	if err != nil {
		return err
	}
	for _, fi := range fis {
		if fi.Mode()&os.ModeSymlink != 0 {
			fi, err = os.Stat(filepath.Join(b.root, fi.Name()))
			if err != nil {
				return err
			}
		}
		if fi.IsDir() {
			b.devices = append(b.devices, &device{
				dir:         filepath.Join(b.root, fi.Name()),
				baseCurrent: defaultBaseCurrent,
				baseMax:     defaultBaseMax,
			})
		}
	}
	if b.devices == nil {
		return errors.New("not found devices in " + b.root)
	}
	return nil
}

func (b *brightness) ListDevices() ([]string, error) {
	err := b.readDevices()
	if err != nil {
		return nil, err
	}
	var list []string
	for _, d := range b.devices {
		list = append(list, filepath.Base(d.dir))
	}
	return list, nil
}

// pick target devices from pattern
func (b *brightness) PickDevice(pat string) error {
	if err := b.readDevices(); err != nil {
		return err
	}
	re, err := regexp.Compile(pat)
	if err != nil {
		return err
	}
	var picked []*device
	for _, d := range b.devices {
		if re.MatchString(filepath.Base(d.dir)) {
			picked = append(picked, d)
		}
	}
	if len(picked) != 1 {
		return errors.New("some devices matched, require matched only one")
	}
	b.picked = picked[0]
	return nil
}

// pick target devices from index
func (b *brightness) PickDeviceIndex(index int) error {
	if err := b.readDevices(); err != nil {
		return err
	}
	if len(b.devices) < index {
		return errors.New("invalid index " + strconv.Itoa(index))
	}
	b.picked = b.devices[index]
	return nil
}

func (b *brightness) Current() (uint, error) {
	if b.picked == nil {
		if err := b.PickDeviceIndex(0); err != nil {
			return 0, err
		}
	}
	return b.picked.current()
}

func (b *brightness) Max() (uint, error) {
	if b.picked == nil {
		if err := b.PickDeviceIndex(0); err != nil {
			return 0, err
		}
	}
	return b.picked.max()
}

func (b *brightness) Set(ui uint) error {
	if b.picked == nil {
		if err := b.PickDeviceIndex(0); err != nil {
			return err
		}
	}
	return b.picked.set(ui)
}
