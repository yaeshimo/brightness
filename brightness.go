// Package brightness provides brightness control.
package brightness

import "errors"

// Append?
//
// list target devices
// ListDevices() []string
//
// specify target device
// PickDevice(target string) error
//
// set directly
// Set(uint) error
//
type Brightness interface {
	// get brightness
	Current() (uint, error)
	Max() (uint, error)

	// set brightness
	Set(uint) error
}

// implement in "brightness_*.go"
var internal Brightness

func Current() (uint, error) { return internal.Current() }
func Max() (uint, error)     { return internal.Max() }

// always returns uint is over the 10
func max() (uint, error) {
	max, err := internal.Max()
	if err != nil {
		return 0, err
	}
	if max < 10 {
		return 0, errors.New("max brightness is too little")
	}
	return max, nil
}

func SetMax() error {
	max, err := internal.Max()
	if err != nil {
		return err
	}
	return internal.Set(max)
}

func SetMid() error {
	max, err := max()
	if err != nil {
		return err
	}
	return internal.Set((max / 10) * 5)
}

func SetMin() error {
	max, err := max()
	if err != nil {
		return err
	}
	return internal.Set(max / 10)
}

// consider:
// setPercent
// (Inc|Dec)10Percent() to (Inc|Dec)Percent(uint)

// if over the max then to max
func Inc10Percent() error {
	max, err := max()
	if err != nil {
		return err
	}
	current, err := internal.Current()
	if err != nil {
		return err
	}
	want := current + (max / 10)
	if want > max {
		want = max
	}
	return internal.Set(want)
}

// if under the 10% then to 10%
func Dec10Percent() error {
	max, err := max()
	if err != nil {
		return err
	}
	current, err := internal.Current()
	if err != nil {
		return err
	}
	tenPercent := max / 10

	// case overflow
	if current < tenPercent {
		return nil
	}

	want := current - tenPercent
	if want < tenPercent {
		want = tenPercent
	}
	return internal.Set(want)
}
