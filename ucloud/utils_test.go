package ucloud

import (
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"
)

func Test_writeToFile(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Error(err)
		return
	}

	type args struct {
		filePath string
		data     interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			"ok1",
			args{path.Join(tempdir, "test_ok.txt"), "str"},
			"str",
			false,
		},
		{
			"ok2",
			args{path.Join(tempdir, "test_ok.txt"), 123},
			"123",
			false,
		},
		{
			"ok3",
			args{path.Join(tempdir, "test_ok.txt"), map[string]string{"foo": "bar"}},
			"{\n\t\"foo\": \"bar\"\n}",
			false,
		},
		{
			"err_empty",
			args{path.Join(tempdir, "test_ok.txt"), ""},
			"",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := writeToFile(tt.args.filePath, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("writeToFile() error = %v, wantErr %v", err, tt.wantErr)
			}

			if isFileExists(tt.args.filePath) {
				t.Errorf("folder is not exists,%v", err)
			}

			if val, err := ioutil.ReadFile(tt.args.filePath); err != nil || string(val) != tt.want {
				t.Errorf("file content = %v, want %v, error %v", string(val), tt.want, err)
			}
		})
	}
}

func isFileExists(filePath string) bool {
	if _, err := os.Stat(filePath); err != nil && os.IsNotExist(err) {
		return true
	}
	return false
}

func Test_buildReversedStringMap(t *testing.T) {
	type args struct {
		input map[string]string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{"ok", args{map[string]string{"key": "value"}}, map[string]string{"value": "key"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildReversedStringMap(tt.args.input); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildReversedStringMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_hashCIDR(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"ok", args{"192.168.0.0/16"}, 494140204},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hashCIDR(tt.args.v); got != tt.want {
				t.Errorf("hashCIDR() = %v, want %v", got, tt.want)
			}
		})
	}
}
