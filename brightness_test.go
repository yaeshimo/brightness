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
		m := &mock{current: test.m.current, max: test.m.max, err: test.m.err}
		internal = m
		if err := f(); err != nil {
			t.Fatal(err)
		}
		out := m.current
		if test.want != out {
			t.Fatalf("case %+v: want %d but out %d\n", test.m, test.want, out)
		}
	}
}

func TestSetMax(t *testing.T) {
	ts := tests{
		{m: &mock{current: 100, max: 100}, want: 100},
		{m: &mock{current: 10, max: 100}, want: 100},
		{m: &mock{current: 1, max: 1}, want: 1},
		{m: &mock{current: 0, max: 1}, want: 1},
	}
	testInternal(t, ts, SetMax)
}

func TestSetMid(t *testing.T) {
	ts := tests{
		{m: &mock{current: 100, max: 100}, want: 50},
		{m: &mock{current: 100, max: 110}, want: 55},
		{m: &mock{current: 111, max: 111}, want: 55},
		{m: &mock{current: 9, max: 9}, want: 4},
		{m: &mock{current: 1, max: 1}, want: 1},
		{m: &mock{current: 0, max: 1}, want: 1},
	}
	testInternal(t, ts, SetMid)
}

func TestSetMin(t *testing.T) {
	ts := tests{
		{m: &mock{current: 100, max: 100}, want: 10},
		{m: &mock{current: 10, max: 10}, want: 1},
		{m: &mock{current: 1, max: 1}, want: 1},
		{m: &mock{current: 0, max: 1}, want: 1},

		// case of max brightness under 10
		{m: &mock{current: 9, max: 9}, want: 1},
	}
	testInternal(t, ts, SetMin)
}

func TestInc10Percent(t *testing.T) {
	ts := tests{
		{m: &mock{current: 100, max: 100}, want: 100},
		{m: &mock{current: 30, max: 300}, want: 60},
		{m: &mock{current: 99, max: 100}, want: 100},
		{m: &mock{current: 1, max: 1}, want: 1},
		{m: &mock{current: 0, max: 1}, want: 1},

		// case of max brightness under 10
		{m: &mock{current: 1, max: 9}, want: 2},
		{m: &mock{current: 9, max: 9}, want: 9},
	}
	testInternal(t, ts, Inc10Percent)
}

func TestDec10Percent(t *testing.T) {
	ts := tests{
		{m: &mock{current: 100, max: 100}, want: 90},
		{m: &mock{current: 30, max: 300}, want: 30},
		{m: &mock{current: 1, max: 1}, want: 1},
		{m: &mock{current: 0, max: 1}, want: 0},

		// case of overflow not change the current
		// change to 10%?
		{m: &mock{current: 10, max: 100}, want: 10},
		{m: &mock{current: 9, max: 100}, want: 9},

		// case of max brightness under 10
		{m: &mock{current: 9, max: 9}, want: 8},
		{m: &mock{current: 1, max: 9}, want: 1},
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
