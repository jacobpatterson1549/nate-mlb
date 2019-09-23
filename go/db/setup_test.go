package db

import (
	"os"
	"testing"
	"time"
)

type mockFileInfo struct {
	NameFunc    func() string
	SizeFunc    func() int64
	ModeFunc    func() os.FileMode
	ModTimeFunc func() time.Time
	IsDirFunc   func() bool
	SysFunc     func() interface{}
}

func (m mockFileInfo) Name() string {
	return m.NameFunc()
}
func (m mockFileInfo) Size() int64 {
	return m.SizeFunc()
}
func (m mockFileInfo) Mode() os.FileMode {
	return m.ModeFunc()
}
func (m mockFileInfo) ModTime() time.Time {
	return m.ModTimeFunc()
}
func (m mockFileInfo) IsDir() bool {
	return m.IsDirFunc()
}
func (m mockFileInfo) Sys() interface{} {
	return m.SysFunc()
}

/*
getSetupFileContents = func(filename string) ([]byte, error) {
	return []byte{}, nil
}
getSetupFunctionDirContents = func(dirname string) ([]os.FileInfo, error) {
	return []os.FileInfo{}, nil
}
*/

func TestPasswordIsValid(t *testing.T) {
	passwordIsValidTests := []struct {
		password password
		want     bool
	}{
		{
			password: "okPassword123",
			want:     true,
		},
		{
			password: "",
			want:     false,
		},
		{
			password: "no spaces are allowed",
			want:     false,
		},
	}
	for i, test := range passwordIsValidTests {
		got := test.password.isValid()
		if test.want != got {
			t.Errorf("Test %d: wanted '%v', but got '%v' for password.isValid() on '%v'", i, test.want, got, test.password)
		}
	}
}
