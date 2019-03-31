package brightness

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"testing"
)

// for Device
type mock struct {
	name             string
	current, max     uint
	cerr, merr, serr error
}

func (m *mock) Name() string           { return m.name }
func (m *mock) Current() (uint, error) { return m.current, m.cerr }
func (m *mock) Max() (uint, error)     { return m.max, m.merr }
func (m *mock) Set(ui uint) error      { m.current = ui; return m.serr }

func (m *mock) String() string {
	return fmt.Sprintf("name:%q current:%d max:%d cerr:%+v merr:%+v serr:%+v",
		m.name, m.current, m.max, m.cerr, m.merr, m.serr)
}

type setTest struct {
	mock    *mock
	want    uint
	wanterr bool
}

func testSet(t *testing.T, ts []setTest, funcName string, force bool) {
	for _, test := range ts {
		// copy
		save := &mock{
			name:    test.mock.name,
			current: test.mock.current, max: test.mock.max,
			cerr: test.mock.cerr, merr: test.mock.merr, serr: test.mock.serr,
		}
		d := Device{max: test.mock.max, internal: test.mock}
		var f interface{}
		switch funcName {
		case "Set":
			f = d.Set
		case "SetMax":
			f = d.SetMax
		case "SetMid":
			f = d.SetMid
		case "SetMin":
			f = d.SetMin
		case "Inc10Percent":
			f = d.Inc10Percent
		case "Dec10Percent":
			f = d.Dec10Percent
		default:
			t.Fatalf("undefined funcName %s", funcName)
		}

		var err error
		switch v := f.(type) {
		case func(uint, bool) error:
			err = v(test.want, force)
		case func() error:
			err = v()
		default:
			t.Fatalf("unexpected test func %s", funcName)
		}

		fatal := func() {
			t.Helper()
			t.Logf("TestFunc:%s case %+v force %v want %d",
				funcName,
				save,
				force,
				test.want)
			t.FailNow()
		}
		if test.wanterr {
			if err != nil {
				continue
			}
			t.Error("expected error but nil")
			fatal()
		}
		if err != nil {
			t.Error(err)
			fatal()
		}
		out := test.mock.current
		if test.want != out {
			t.Errorf("unexpected output %d", out)
			fatal()
		}
	}
}

func testSetFunc(t *testing.T, ts []setTest, funcName string) {
	testSet(t, ts, funcName, false)
}

type setData struct {
	current, max uint
	want         uint
}

func makeTests(sd []setData) []setTest {
	var ts []setTest
	for _, d := range sd {
		ts = append(ts, setTest{
			mock: &mock{current: d.current, max: d.max},
			want: d.want,
		})
	}
	return ts
}

type errorData struct {
	current, max     uint
	want             uint
	cerr, merr, serr bool
}

func makeErrorTests(ed []errorData) []setTest {
	var ts []setTest
	for _, d := range ed {
		m := &mock{current: d.current, max: d.max}
		if d.cerr {
			m.cerr = errors.New("error from current")
		}
		if d.merr {
			m.merr = errors.New("error from max")
		}
		if d.serr {
			m.serr = errors.New("error from set")
		}
		ts = append(ts, setTest{
			mock:    m,
			want:    d.want,
			wanterr: true,
		})
	}
	return ts
}

var setTests = []setData{
	{100, 100, 100},
	{10, 100, 50},
	{1, 1, 1},
	{0, 1, 1},
}

var setForceTests = []setData{
	{100, 100, 0},
	{100, 100, 9},
	{1, 1, 0},
}

var setErrorTests = []errorData{
	// over the max
	{100, 100, 101, false, false, false},
	{1, 1, 2, false, false, false},

	// under limit
	{10, 100, 0, false, false, false},
	{10, 100, 9, false, false, false},
	{1, 1, 0, false, false, false},

	// internal error
	{100, 100, 100, false, false, true},
}

var setForceErrorTests = []errorData{
	// over the max
	{100, 100, 101, false, false, false},
	{1, 1, 2, false, false, false},

	// internal error
	{100, 100, 100, false, false, true},
}

func TestSet(t *testing.T) {
	testSet(t, makeTests(setTests), "Set", false)
	testSet(t, makeErrorTests(setErrorTests), "Set", false)

	testSet(t, makeTests(setTests), "Set", true)
	testSet(t, makeErrorTests(setForceErrorTests), "Set", true)
}

var setMaxTests = []setData{
	{100, 100, 100},
	{50, 100, 100},
	{10, 100, 100},
	{0, 100, 100},
	{1, 1, 1},
	{0, 1, 1},
}

func TestSetMax(t *testing.T) {
	testSetFunc(t, makeTests(setMaxTests), "SetMax")
}

var setMidTests = []setData{
	{100, 100, 50},
	{100, 110, 55},
	{111, 111, 55},
	{0, 100, 50},

	{10, 10, 5},
	{9, 9, 4},

	{1, 1, 1},
	{0, 1, 1},
}

func TestSetMid(t *testing.T) {
	testSetFunc(t, makeTests(setMidTests), "SetMid")
}

var setMinTests = []setData{
	{100, 100, 10},
	{10, 10, 1},
	{9, 9, 1},
	{1, 1, 1},

	// need change?
	{0, 100, 10},
	{0, 1, 1},
}

func TestSetMin(t *testing.T) {
	testSetFunc(t, makeTests(setMinTests), "SetMin")
}

var inc10PercentTests = []setData{
	{100, 100, 100},
	{9, 9, 9},
	{1, 1, 1},

	{80, 100, 90},
	{1, 9, 2},
	{0, 1, 1},

	// need error?
	{100, 90, 90},
}

var inc10PercentErrorTests = []errorData{
	// internal error
	{100, 100, 100, true, false, false},
}

func TestInc10Percent(t *testing.T) {
	testSetFunc(t, makeTests(inc10PercentTests), "Inc10Percent")
	testSetFunc(t, makeErrorTests(inc10PercentErrorTests), "Inc10Percent")
}

var dec10PercentTests = []setData{
	{100, 100, 90},
	{10, 10, 9},
	{9, 10, 8},
	{9, 9, 8},

	// need error?
	{200, 100, 190},
	{9, 5, 8},

	// limit
	{10, 100, 10},
	{19, 100, 10},
	{1, 1, 1},
	{1, 9, 1},

	// case of already under the limits
	// change to 10 or 1?
	{9, 100, 9},
	{0, 1, 0},
}

var dec10PercentErrorTests = []errorData{
	// internal error
	{100, 100, 90, true, false, false},
}

func TestDec10Percent(t *testing.T) {
	testSetFunc(t, makeTests(dec10PercentTests), "Dec10Percent")
	testSetFunc(t, makeErrorTests(dec10PercentErrorTests), "Dec10Percent")
}

func TestBrightnessState(t *testing.T) {
	device := Device{
		max: 100,
		internal: &mock{
			name:    "mock device",
			current: 100,
			max:     100,
		}}

	t.Logf("name %s", device.Name())
	ui, err := device.Current()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("current brightness: %d", ui)
	t.Logf("max brightness: %d", device.Max())
}

var readDeviceAllTests = []struct {
	mocks   []*mock
	err     error // internal error
	wanterr bool
}{
	// valid
	{
		mocks: []*mock{
			{name: "mock", current: 100, max: 100},
		},
	},
	{
		mocks: []*mock{
			{name: "mock1", current: 100, max: 100},
			{name: "mock2", current: 100, max: 100},
		},
	},

	// return error from readDeviceAll
	{
		mocks: []*mock{
			{name: "mock", current: 100, max: 100},
		},
		err:     errors.New("internal error"),
		wanterr: true,
	},

	// not found devices
	{
		wanterr: true,
	},

	// max brightness is 0
	{
		mocks: []*mock{
			{name: "mock", current: 100, max: 0},
		},
		wanterr: true,
	},

	// return error from internal.Max()
	{
		mocks: []*mock{
			{name: "mock", current: 100, max: 100, merr: errors.New("error Max()")},
		},
		wanterr: true,
	},

	// duplicated device names
	{
		mocks: []*mock{
			{name: "mock", current: 100, max: 100},
			{name: "mock", current: 100, max: 100},
		},
		wanterr: true,
	},
}

var readDeviceIndexTests = []struct {
	mocks []*mock
	index int
	exp   *mock

	err     error // internal error
	wanterr bool
}{
	// valid
	{
		mocks: []*mock{
			{name: "mock", current: 100, max: 100},
		},
		index: 0,
		exp:   &mock{name: "mock", current: 100, max: 100},
	},
	{
		mocks: []*mock{
			{name: "mock1", current: 100, max: 100},
			{name: "mock2", current: 100, max: 100},
		},
		index: 1,
		exp:   &mock{name: "mock2", current: 100, max: 100},
	},

	// internal error
	{
		mocks: []*mock{
			{name: "mock", current: 100, max: 100},
		},
		index:   0,
		err:     errors.New("internal error"),
		wanterr: true,
	},

	// out of length
	{
		mocks: []*mock{
			{name: "mock", current: 100, max: 100},
		},
		index:   1,
		wanterr: true,
	},
	{
		mocks: []*mock{
			{name: "mock", current: 100, max: 100},
		},
		index:   -1,
		wanterr: true,
	},
}

var readDevicePatTests = []struct {
	mocks []*mock
	pat   string
	exp   []*mock

	err     error // internal error
	wanterr bool
}{
	// valid
	{
		mocks: []*mock{
			{name: "mock", current: 100, max: 100},
		},
		pat: ".*",
		exp: []*mock{
			{name: "mock", current: 100, max: 100},
		},
	},
	{
		mocks: []*mock{
			{name: "mock1", current: 100, max: 100},
			{name: "mock2", current: 100, max: 100},
		},
		pat: ".*",
		exp: []*mock{
			{name: "mock1", current: 100, max: 100},
			{name: "mock2", current: 100, max: 100},
		},
	},
	{
		mocks: []*mock{
			{name: "mock1", current: 100, max: 100},
			{name: "mock2", current: 100, max: 100},
			{name: "prefix-mock3", current: 100, max: 100},
		},
		pat: "^prefix-.*$",
		exp: []*mock{
			{name: "prefix-mock3", current: 100, max: 100},
		},
	},

	// invalid pattern
	{
		mocks: []*mock{
			{name: "mock", current: 100, max: 100},
		},
		pat:     "*",
		wanterr: true,
	},

	// internal error
	{
		mocks: []*mock{
			{name: "mock", current: 100, max: 100},
		},
		pat:     ".*",
		err:     errors.New("internal error"),
		wanterr: true,
	},

	// not match
	{
		mocks: []*mock{
			{name: "mock", current: 100, max: 100},
		},
		pat:     "^not match$",
		wanterr: true,
	},
}

func TestReadDevice(t *testing.T) {
	tmpf := readDeviceAll
	defer func() { readDeviceAll = tmpf }()

	initFunc := func(mocks []*mock, err error) {
		readDeviceAll = func() ([]*Device, error) {
			devices := make([]*Device, 0, len(mocks))
			for _, m := range mocks {
				devices = append(devices, &Device{
					internal: m,
					max:      m.max,
				})
			}
			return devices, err
		}
	}
	log := func(ds []*Device) string {
		var s string
		for _, d := range ds {
			s += fmt.Sprintf("%+v %+v %+v",
				d.Name(), fmt.Sprint(d.Current()), fmt.Sprint(d.Max()))
		}
		return s
	}
	verify := func(exp []*Device, out []*Device) bool {
		t.Helper()
		sort.Slice(exp, func(i, j int) bool { return exp[i].Name() < exp[j].Name() })
		if !reflect.DeepEqual(exp, out) {
			t.Error("unexpected out")
			t.Errorf("exp:\n%s", log(exp))
			t.Errorf("out:\n%s", log(out))
			return false
		}
		return true
	}

	t.Run("ReadDeviceAll", func(t *testing.T) {
		for _, test := range readDeviceAllTests {
			initFunc(test.mocks, test.err)
			out, err := ReadDeviceAll()
			if test.wanterr {
				if err != nil {
					continue
				}
				t.Errorf("case %+v\nmocks:%+v", test, test.mocks)
				t.Fatal("expected error but nil")
			}
			if err != nil {
				t.Fatal(err)
			}

			exp, _ := readDeviceAll()
			if !verify(exp, out) {
				t.Errorf("case %+v", test)
				var s string
				for _, m := range test.mocks {
					s += fmt.Sprint(m)
				}
				t.Fatalf("mocks:\n%s", s)
			}
		}
	})

	t.Run("ReadDeviceIndex", func(t *testing.T) {
		for _, test := range readDeviceIndexTests {
			initFunc(test.mocks, test.err)
			out, err := ReadDeviceIndex(test.index)
			if test.wanterr {
				if err != nil {
					continue
				}
				t.Errorf("case %+v\nmocks:%+v", test, test.mocks)
				t.Fatal("expected error but nil")
			}
			if err != nil {
				t.Fatal(err)
			}
			exp := []*Device{
				{
					internal: test.exp,
					max:      test.exp.max,
				},
			}
			if !verify(exp, []*Device{out}) {
				t.Errorf("case %+v", test)
				var s string
				for _, m := range test.mocks {
					s += fmt.Sprint(m)
				}
				t.Fatalf("mocks:\n%s", s)
			}
		}
	})

	t.Run("ReadDevicePat", func(t *testing.T) {
		for _, test := range readDevicePatTests {
			initFunc(test.mocks, test.err)
			out, err := ReadDevicePat(test.pat)
			if test.wanterr {
				if err != nil {
					continue
				}
				t.Errorf("case %+v\nmocks:%+v", test, test.mocks)
				t.Fatal("expected error but nil")
			}
			if err != nil {
				t.Fatal(err)
			}
			exp := make([]*Device, 0, len(test.exp))
			for _, i := range test.exp {
				exp = append(exp, &Device{
					internal: i,
					max:      i.max,
				})
			}
			if !verify(exp, out) {
				t.Errorf("case %+v", test)
				var s string
				for _, m := range test.mocks {
					s += fmt.Sprint(m)
				}
				t.Fatalf("mocks:\n%s", s)
			}
		}
	})
}
