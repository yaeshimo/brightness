// +build linux

package brightness

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// implement for Brightness
//
// expected locations
// devices : "/sys/class/backlight/"
// device  : "/sys/class/backlight/*/"
// files   : "/sys/class/backlight/*/{max_,}brightness
type brightness struct {
	// "/sys/class/backlight"
	root string

	// TODO: change to full path?
	// base "brightness"
	current string

	// TODO: change to full path?
	// base "max_brightness"
	max string

	// pick one from "brightness.root/*"
	device string
}

func init() {
	internal = &brightness{
		root:    "/sys/class/backlight",
		current: "brightness",
		max:     "max_brightness",
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

// from "b.root/*"
func (b *brightness) candidates() ([]string, error) {
	var candidates []string
	fis, err := ioutil.ReadDir(b.root)
	if err != nil {
		return nil, err
	}
	for _, fi := range fis {
		if fi.Mode()&os.ModeSymlink != 0 {
			fi, err = os.Stat(filepath.Join(b.root, fi.Name()))
			if err != nil {
				return nil, err
			}
		}
		if fi.IsDir() {
			candidates = append(candidates, filepath.Join(b.root, fi.Name()))
		}
	}
	return candidates, nil
}

// TODO: consider what pick one
// add arguments for to select device
func (b *brightness) pickTarget() error {
	// TODO: fix
	// if already called then do not anything
	if b.device != "" {
		return nil
	}
	dirs, err := b.candidates()
	if err != nil {
		return err
	}
	switch len(dirs) {
	case 0:
		return fmt.Errorf("read %q but not found candidate directories", b.root)
	case 1:
		b.device = dirs[0]
	default:
		// TODO: fix
		b.device = dirs[0]
	}
	return nil
}

func (b *brightness) Current() (uint, error) {
	if err := b.pickTarget(); err != nil {
		return 0, err
	}
	return readUint(filepath.Join(b.device, b.current))
}

func (b *brightness) Max() (uint, error) {
	if err := b.pickTarget(); err != nil {
		return 0, err
	}
	return readUint(filepath.Join(b.device, b.max))
}

func (b *brightness) Set(ui uint) error {
	if err := b.pickTarget(); err != nil {
		return err
	}

	max, err := b.Max()
	if err != nil {
		return err
	}
	if ui > max {
		return errors.New("requested brightness is over the max")
	}
	f, err := os.OpenFile(filepath.Join(b.device, b.current), os.O_WRONLY|os.O_TRUNC, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(strconv.FormatUint(uint64(ui), 10))
	return err
}
