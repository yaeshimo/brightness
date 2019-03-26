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
		current, max uint
		set          uint
	}{
		{100, 100, 30},
		{100, 100, 10},
		{9, 9, 8},

		// permit small set
		{100, 100, 0},
		{100, 100, 1},
		{9, 9, 0},

		// want error
		{100, 100, 101},
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
		b := &brightness{root: rootdir}
		writeUint(subdir, defaultBaseCurrent, test.current)
		writeUint(subdir, defaultBaseMax, test.max)

		if err := b.Set(test.set); err != nil {
			// case want error
			if test.set > test.max {
				continue
			}
			t.Fatal(err)
		} else if test.set > test.max {
			t.Fatal("want error but nil")
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
