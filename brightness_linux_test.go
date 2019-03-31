// +build linux

package brightness

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func writeFile(dir, base string, content string) error {
	return ioutil.WriteFile(filepath.Join(dir, base), []byte(content), 0600)
}

func makeDeviceDir(dir string, current, max string) (string, error) {
	deviceRoot, err := ioutil.TempDir(dir, "")
	if err != nil {
		return "", err
	}
	err = writeFile(deviceRoot, baseCurrent, current)
	if err != nil {
		return "", err
	}
	err = writeFile(deviceRoot, baseMax, max)
	if err != nil {
		return "", err
	}
	return deviceRoot, nil
}

var linuxSetTests = []struct {
	current, max string
	want         uint
	wanterr      bool
}{
	{"100", "100", 100, false},
	{"100", "100", 50, false},
	{"100", "100", 10, false},
	{"0", "100", 1, false},
	{"9", "9", 8, false},
	{"9", "9", 1, false},
	{"0", "9", 1, false},

	// permit the small number
	{"100", "100", 0, false},
	{"100", "100", 1, false},
	{"9", "9", 0, false},

	// want error
	{"100", "100", 101, true},
	{"100", "string", 100, true},

	// want error?
	//{"string", "100", 100, true},
}

func TestSet_Linux(t *testing.T) {
	// expected locations
	// root       : "/sys/class/backlight/"
	// deviceRoot : "/sys/class/backlight/*/"
	// files      : "/sys/class/backlight/*/{max_,}brightness
	testRoot, err := ioutil.TempDir("", "TestSet_Linux")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testRoot)

	for _, test := range linuxSetTests {
		deviceRoot, err := makeDeviceDir(testRoot, test.current, test.max)
		if err != nil {
			t.Fatal(err)
		}
		d := &device{root: deviceRoot}
		err = d.Set(test.want)
		if test.wanterr {
			if err != nil {
				continue
			}
			t.Fatalf("case %+v expected error but nil", test)
		}
		if err != nil {
			t.Fatalf("case %+v %v", test, err)
		}
		out, err := d.Current()
		if err != nil {
			t.Fatalf("case %+v %v", test, err)
		}
		if test.want != out {
			t.Fatalf("want %d but out %d", test.want, out)
		}
	}

	t.Run("Case Rejected", func(t *testing.T) {
		deviceRoot, err := makeDeviceDir(testRoot, "100", "100")
		if err != nil {
			t.Fatal(err)
		}
		d := &device{root: deviceRoot}
		err = os.RemoveAll(deviceRoot)
		if err != nil {
			t.Fatal(err)
		}
		err = d.Set(50)
		if err == nil {
			t.Fatal("expected error but nil")
		}
	})

	t.Run("Case Missing Files", func(t *testing.T) {
		for _, file := range []string{baseCurrent, baseMax} {
			deviceRoot, err := makeDeviceDir(testRoot, "100", "100")
			if err != nil {
				t.Fatal(err)
			}
			d := &device{root: deviceRoot}
			err = os.Remove(filepath.Join(deviceRoot, file))
			if err != nil {
				t.Fatal(err)
			}
			err = d.Set(50)
			if err == nil {
				t.Fatal("expected error but nil")
			}
		}
	})
}

func TestReadDevice_Linux(t *testing.T) {
	testRoot, err := ioutil.TempDir("", "TestReadDevice_Linux")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testRoot)

	verify := func(t *testing.T, classRoot string, exp []*Device, wanterr bool) {
		t.Helper()
		tmp := root
		defer func() { root = tmp }()
		root = classRoot
		out, err := readDeviceAll()
		if wanterr {
			if err != nil {
				return
			}
			t.Fatal("expected error but nil")
		}
		if err != nil {
			t.Fatal(err)
		}
		// expected devices always sorted
		sort.Slice(exp, func(i, j int) bool { return exp[i].Name() < exp[j].Name() })
		if !reflect.DeepEqual(exp, out) {
			t.Errorf("unexpected output, exp != out")
			t.Log("exp")
			for i := range exp {
				t.Logf("\t%+v", exp[i])
			}
			t.Log("out")
			for i := range out {
				t.Logf("\t%+v", out[i])
			}
			t.FailNow()
		}
	}

	t.Run("Read One", func(t *testing.T) {
		classRoot, err := ioutil.TempDir(testRoot, "")
		if err != nil {
			t.Fatal(err)
		}
		deviceRoot, err := makeDeviceDir(classRoot, "100", "100")
		if err != nil {
			t.Fatal(err)
		}
		exp := []*Device{
			{
				internal: &device{root: deviceRoot},
				max:      100,
			},
		}
		verify(t, classRoot, exp, false)
	})

	t.Run("Read Symlink", func(t *testing.T) {
		classRoot, err := ioutil.TempDir(testRoot, "")
		if err != nil {
			t.Fatal(err)
		}
		deviceRoot, err := makeDeviceDir(classRoot, "100", "100")
		if err != nil {
			t.Fatal(err)
		}
		outOfClass, err := makeDeviceDir(testRoot, "100", "100")
		if err != nil {
			t.Fatal(err)
		}
		sym := filepath.Join(classRoot, filepath.Base(outOfClass))
		err = os.Symlink(outOfClass, sym)
		if err != nil {
			t.Fatal(err)
		}
		exp := []*Device{
			{
				internal: &device{deviceRoot},
				max:      100,
			},
			{
				internal: &device{sym},
				max:      100,
			},
		}
		verify(t, classRoot, exp, false)
	})

	// want error

	t.Run("Not Found Devices Root", func(t *testing.T) {
		classRoot, err := ioutil.TempDir(testRoot, "")
		if err != nil {
			t.Fatal(err)
		}
		err = os.Remove(classRoot)
		if err != nil {
			t.Fatal(err)
		}
		verify(t, classRoot, nil, true)
	})

	t.Run("Not Found Devices Dir", func(t *testing.T) {
		classRoot, err := ioutil.TempDir(testRoot, "")
		if err != nil {
			t.Fatal(err)
		}
		verify(t, classRoot, nil, true)
	})

	t.Run("Invalid Symlink", func(t *testing.T) {
		classRoot, err := ioutil.TempDir(testRoot, "")
		if err != nil {
			t.Fatal(err)
		}
		outOfClass, err := makeDeviceDir(testRoot, "100", "100")
		if err != nil {
			t.Fatal(err)
		}
		invalidSym := filepath.Join(classRoot, filepath.Base(outOfClass))
		err = os.Symlink(outOfClass, invalidSym)
		if err != nil {
			t.Fatal(err)
		}
		err = os.RemoveAll(outOfClass)
		if err != nil {
			t.Fatal(err)
		}
		verify(t, classRoot, nil, true)
	})

	t.Run("Can Not Read Max", func(t *testing.T) {
		classRoot, err := ioutil.TempDir(testRoot, "")
		if err != nil {
			t.Fatal(err)
		}
		deviceRoot, err := makeDeviceDir(classRoot, "100", "100")
		if err != nil {
			t.Fatal(err)
		}
		err = os.Remove(filepath.Join(deviceRoot, baseMax))
		if err != nil {
			t.Fatal(err)
		}
		verify(t, classRoot, nil, true)
	})
}
