// +build linux

package brightness

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestSetBrightness(t *testing.T) {
	tmpRoot, err := ioutil.TempDir("", "test_brightness")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpRoot)

	writeUint := func(dir, base string, ui uint) {
		b := []byte(strconv.FormatUint(uint64(ui), 10))
		err := ioutil.WriteFile(filepath.Join(dir, base), b, 0600)
		if err != nil {
			t.Fatal(err)
		}
	}

	var tests = []struct {
		current uint
		max     uint
		set     uint
	}{
		{current: 100, max: 100, set: 30},
		{current: 100, max: 100, set: 10},
		{current: 9, max: 9, set: 8},

		// permit?
		{current: 100, max: 100, set: 0},
		{current: 100, max: 100, set: 1},
		{current: 9, max: 9, set: 0},

		// want error
		{current: 100, max: 100, set: 101},
	}

	for _, test := range tests {
		// expected locations
		// rootdir: "/sys/class/backlight/"
		// subdir : "/sys/class/backlight/*/"
		// files  : "/sys/class/backlight/*/{max_,}brightness
		rootdir, err := ioutil.TempDir(tmpRoot, "")
		if err != nil {
			t.Fatal(err)
		}
		subdir, err := ioutil.TempDir(rootdir, "")
		if err != nil {
			t.Fatal(err)
		}
		b := &brightness{
			root:    rootdir,
			current: "brightness",
			max:     "max_brightness",
		}
		writeUint(subdir, b.current, test.current)
		writeUint(subdir, b.max, test.max)

		if err := b.Set(test.set); err != nil {
			// case of want error
			if test.set > test.max {
				continue
			}
			t.Fatal(err)
		}
		out, err := b.Current()
		if err != nil {
			t.Fatal(err)
		}
		if test.set != out {
			t.Fatalf("want %d but out %d", test.set, out)
		}
	}
}
