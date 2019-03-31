// package brightness provides brightness control.
package brightness

import (
	"errors"
	"regexp"
	"sort"
	"strconv"
)

// implement in brightness_*.go
type internal interface {
	// expected name is always unique
	Name() string
	Current() (uint, error)
	Max() (uint, error)
	Set(uint) error
}

type Device struct {
	internal internal

	// TODO: remove? to use Max()(uint, error)?
	// expected max is always greater than 1
	max uint
}

// implement in brightness_*.go
var readDeviceAll func() ([]*Device, error)

func ReadDeviceAll() ([]*Device, error) {
	devices, err := readDeviceAll()
	if err != nil {
		return nil, err
	}
	if len(devices) == 0 {
		return nil, errors.New("can not found devices")
	}
	duplicate := make(map[string]bool, len(devices))
	for i := range devices {
		max, err := devices[i].internal.Max()
		if err != nil {
			return nil, err
		}
		if max == 0 {
			return nil, errors.New(devices[i].Name() + " max brightness is 0")
		}
		devices[i].max = max
		if name := devices[i].internal.Name(); duplicate[name] {
			return nil, errors.New("device name " + name + " is duplicated")
		} else {
			duplicate[name] = true
		}
	}
	sort.Slice(devices, func(i, j int) bool {
		return devices[i].Name() < devices[j].Name()
	})
	return devices, nil
}

// TODO: is need?
// change to func(index ...int) ([]*Device, error)?
func ReadDeviceIndex(index int) (*Device, error) {
	devices, err := ReadDeviceAll()
	if err != nil {
		return nil, err
	}
	if n := len(devices) - 1; index < 0 || index > n {
		return nil, errors.New("invalid index " + strconv.Itoa(index))
	}
	return devices[index], err
}

// TODO: is need?
// pick target devices from pattern
func ReadDevicePat(pat string) ([]*Device, error) {
	re, err := regexp.Compile(pat)
	if err != nil {
		return nil, err
	}
	devices, err := ReadDeviceAll()
	if err != nil {
		return nil, err
	}
	var picked []*Device
	for _, d := range devices {
		if re.MatchString(d.Name()) {
			picked = append(picked, d)
		}
	}
	if len(picked) == 0 {
		return nil, errors.New("not match any devices")
	}
	return picked, nil
}

func (d *Device) Name() string           { return d.internal.Name() }
func (d *Device) Current() (uint, error) { return d.internal.Current() }

func (d *Device) Max() uint { return d.max }
func (d *Device) Mid() uint {
	if d.max == 1 {
		return 1
	}
	return d.max / 2
}
func (d *Device) Min() uint {
	if d.max < 10 {
		return 1
	}
	return d.max / 10
}

// if force is true then ignore the limit
func (d *Device) Set(want uint, force bool) error {
	if want > d.max {
		return errors.New("requested brightness over the max")
	}
	if force {
		return d.internal.Set(want)
	}
	if want == 0 {
		return errors.New("can not set brightness to 0")
	}
	if d.max > 10 && want < d.max/10 {
		return errors.New("can not set brightness under the 10 percent")
	}
	return d.internal.Set(want)
}

func (d *Device) SetMax() error { return d.internal.Set(d.max) }
func (d *Device) SetMid() error { return d.internal.Set(d.Mid()) }
func (d *Device) SetMin() error { return d.internal.Set(d.Min()) }

// provide?: SetPercent(i int) error

func (d *Device) Inc10Percent() error {
	current, err := d.internal.Current()
	if err != nil {
		return err
	}
	var want uint
	if d.max < 10 {
		want = current + 1
	} else {
		want = current + d.max/10
	}
	if want > d.max {
		want = d.max
	}
	return d.internal.Set(want)
}

func (d *Device) Dec10Percent() error {
	current, err := d.internal.Current()
	if err != nil {
		return err
	}
	if current <= 1 {
		return nil
	}
	var want uint
	if d.max < 10 {
		want = current - 1
	} else {
		tenPercent := d.max / 10
		if current <= tenPercent {
			return nil
		}
		want = current - tenPercent
		if want < tenPercent {
			want = tenPercent
		}
	}
	return d.internal.Set(want)
}
