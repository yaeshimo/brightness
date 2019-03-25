package brightness

// TODO: remove max brihtness is too small

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

func SetMax() error {
	max, err := internal.Max()
	if err != nil {
		return err
	}
	return internal.Set(max)
}

func SetMid() error {
	max, err := internal.Max()
	if err != nil {
		return err
	}
	want := max / 2
	if want == 0 && max == 1 {
		want = 1
	}
	return internal.Set(want)
}

func SetMin() error {
	max, err := internal.Max()
	if err != nil {
		return err
	}
	if max < 10 {
		return internal.Set(1)
	}
	return internal.Set(max / 10)
}

// consider:
// setPercent
// (Inc|Dec)10Percent() to (Inc|Dec)Percent(uint)

func Inc10Percent() error {
	max, err := internal.Max()
	if err != nil {
		return err
	}
	current, err := internal.Current()
	if err != nil {
		return err
	}
	var want uint
	if max < 10 {
		want = current + 1
	} else {
		want = current + max/10
	}
	if want > max {
		want = max
	}
	return internal.Set(want)
}

func Dec10Percent() error {
	max, err := internal.Max()
	if err != nil {
		return err
	}
	current, err := internal.Current()
	if err != nil {
		return err
	}
	var want uint
	if max < 10 {
		if current <= 1 {
			return nil
		}
		want = current - 1
	} else {
		tenPercent := max / 10
		if current <= tenPercent {
			return nil
		}
		want = current - tenPercent
	}
	return internal.Set(want)
}
