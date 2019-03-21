package brightness

import (
	"testing"
)

// for internal
type mock struct {
	current uint
	max     uint
	err     error
}

// implement Brightness
func (m *mock) Current() (uint, error) { return m.current, m.err }
func (m *mock) Max() (uint, error)     { return m.max, m.err }
func (m *mock) Set(ui uint) error      { m.current = ui; return m.err }

type tests []struct {
	m    *mock
	want uint
}

func testInternal(t *testing.T, ts tests, f func() error) {
	tmpInternal := internal
	defer func() { internal = tmpInternal }()
	t.Helper()
	for _, test := range ts {
		internal = test.m
		err := f()
		if err != nil {
			// case too little max brightness
			if test.m.max < 10 {
				continue
			}
			t.Fatal(err)
		}
		out := test.m.current
		if test.want != out {
			t.Fatalf("expected %d but out %d\n", test.want, out)
		}
	}
}

func TestSetMax(t *testing.T) {
	ts := tests{
		{m: &mock{current: 100, max: 100}, want: 100},
		{m: &mock{current: 10, max: 100}, want: 100},
	}
	testInternal(t, ts, SetMax)
}

func TestSetMid(t *testing.T) {
	ts := tests{
		{m: &mock{current: 100, max: 100}, want: 50},
		{m: &mock{current: 100, max: 110}, want: 55},
		{m: &mock{current: 111, max: 111}, want: 55},
	}
	testInternal(t, ts, SetMid)
}

func TestSetMin(t *testing.T) {
	ts := tests{
		{m: &mock{current: 100, max: 100}, want: 10},
	}
	testInternal(t, ts, SetMin)
}

// white box testing

func TestInc10Percent(t *testing.T) {
	ts := tests{
		{m: &mock{current: 100, max: 100}, want: 100},
		{m: &mock{current: 30, max: 300}, want: 60},
		{m: &mock{current: 99, max: 100}, want: 100},

		// want error
		{m: &mock{current: 9, max: 9}},
	}
	testInternal(t, ts, Inc10Percent)
}

func TestDec10Percent(t *testing.T) {
	ts := tests{
		{m: &mock{current: 100, max: 100}, want: 90},
		{m: &mock{current: 30, max: 300}, want: 30},

		// case overflow not change current
		// change to 10%?
		{m: &mock{current: 10, max: 100}, want: 10},

		// want error
		{m: &mock{current: 9, max: 9}},
	}
	testInternal(t, ts, Dec10Percent)
}

func TestBrightnessState(t *testing.T) {
	tmpInternal := internal
	defer func() { internal = tmpInternal }()
	internal = &mock{
		current: 100,
		max:     100,
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
