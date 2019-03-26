package brightness

import (
	"errors"
	"regexp"
	"testing"
)

type dammyDevice struct {
	// name of device
	name string
	// state of brightness
	current uint
	max     uint
}

// for internal
type mock struct {
	// mock devices
	devices []*dammyDevice
	// picked device
	picked *dammyDevice
	// expected error
	err error
}

// implement Brightness
func (m *mock) Current() (uint, error) { return m.picked.current, m.err }
func (m *mock) Max() (uint, error)     { return m.picked.max, m.err }
func (m *mock) Set(ui uint) error      { m.picked.current = ui; return m.err }

// TODO: implement and test
func (m *mock) ListDevices() ([]string, error) {
	var names []string
	for i := range m.devices {
		names = append(names, m.devices[i].name)
	}
	return names, m.err
}
func (m *mock) PickDevice(pat string) error {
	re, err := regexp.Compile(pat)
	if err != nil {
		return err
	}
	names, err := m.ListDevices()
	if err != nil {
		return err
	}
	var matched []string
	for i := range names {
		if re.MatchString(names[i]) {
			matched = append(matched, names[i])
		}
	}
	if len(matched) != 1 {
		return errors.New("some devices matched, require matched only one")
	}
	return m.err
}
func (m *mock) PickDeviceIndex(index int) error {
	names, err := m.ListDevices()
	if err != nil {
		return err
	}
	if len(names) < index {
		return errors.New("requested index is over the length")
	}
	m.picked = m.devices[index]
	return m.err
}

type setBrightnessTest struct {
	current, max uint
	want         uint
}

func testSetBrightness(t *testing.T, ts []setBrightnessTest, f func() error) {
	tmpInternal := internal
	defer func() { internal = tmpInternal }()
	t.Helper()
	for _, test := range ts {
		m := &mock{picked: &dammyDevice{current: test.current, max: test.max}}
		internal = m
		if err := f(); err != nil {
			t.Fatal(err)
		}
		out := m.picked.current
		if test.want != out {
			t.Fatalf("case %+v: want %d but out %d\n", test, test.want, out)
		}
	}
}

var setMaxTests = []setBrightnessTest{
	{100, 100, 100},
	{10, 100, 100},
	{1, 1, 1},
	{0, 1, 1},
}

func TestSetMax(t *testing.T) {
	testSetBrightness(t, setMaxTests, SetMax)
}

var setMidTests = []setBrightnessTest{
	{100, 100, 50},
	{100, 110, 55},
	{111, 111, 55},
	{9, 9, 4},
	{1, 1, 1},
	{0, 1, 1},
}

func TestSetMid(t *testing.T) {
	testSetBrightness(t, setMidTests, SetMid)
}

var setMinTests = []setBrightnessTest{
	{100, 100, 10},
	{10, 10, 1},
	{1, 1, 1},
	{0, 1, 1},

	// case of max brightness under 10
	{9, 9, 1},
}

func TestSetMin(t *testing.T) {
	testSetBrightness(t, setMinTests, SetMin)
}

var inc10PercentTests = []setBrightnessTest{
	{100, 100, 100},
	{30, 300, 60},
	{99, 100, 100},
	{1, 1, 1},
	{0, 1, 1},

	// case of max brightness under 10
	{1, 9, 2},
	{9, 9, 9},
}

func TestInc10Percent(t *testing.T) {
	testSetBrightness(t, inc10PercentTests, Inc10Percent)
}

var dec10PercentTests = []setBrightnessTest{
	{100, 100, 90},
	{30, 300, 30},
	{1, 1, 1},
	{0, 1, 0},

	// case of overflow not change the current
	// change to 10%?
	{10, 100, 10},
	{9, 100, 9},

	// case of max brightness under 10
	{9, 9, 8},
	{1, 9, 1},
}

func TestDec10Percent(t *testing.T) {
	testSetBrightness(t, dec10PercentTests, Dec10Percent)
}

// TODO: fix
func TestBrightnessState(t *testing.T) {
	tmpInternal := internal
	defer func() { internal = tmpInternal }()
	internal = &mock{
		picked: &dammyDevice{
			current: 100,
			max:     100,
		},
	}

	ui, err := internal.Current()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("current brightness: %d", ui)

	ui, err = internal.Max()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("max brightness: %d", ui)
}
