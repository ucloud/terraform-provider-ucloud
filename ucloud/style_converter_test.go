package ucloud

import (
	"testing"
)

func Test_upperConverter_convert(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name    string
		cvt     styleConverter
		args    args
		want    string
		wantErr bool
	}{
		{"upper", upperCvt, args{"LOCAL_SSD"}, "local_ssd", false},
		{"mix", upperCvt, args{"LoCal_ssd"}, "", true},
		{"noSpan", upperCvt, args{"LOCALSSD"}, "localssd", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.cvt.convertWithErr(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("upperConverter.convertWithErr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("upperConverter.convertWithErr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_upperConverter_unconvert(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name    string
		cvt     styleConverter
		args    args
		want    string
		wantErr bool
	}{
		{"lower", upperCvt, args{"local_ssd"}, "LOCAL_SSD", false},
		{"mix", upperCvt, args{"LoCal_SSD"}, "", true},
		{"noSpan", upperCvt, args{"localssd"}, "LOCALSSD", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.cvt.unconvertWithErr(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("upperConverter.unconvertWithErr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("upperConverter.unconvertWithErr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_lowerCamelConverter_convert(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name    string
		cvt     styleConverter
		args    args
		want    string
		wantErr bool
	}{
		{"lower", lowerCamelCvt, args{"success"}, "success", false},
		{"lowerCamel", lowerCamelCvt, args{"createFail"}, "create_fail", false},
		{"lowerCamelWithSpecial", lowerCamelCvt, args{"createUDBFail"}, "create_udb_fail", false},
		{"lowerCamelWithSpecial", lowerCamelCvt, args{"localSSD"}, "local_ssd", false},
		{"upper", lowerCamelCvt, args{"HA"}, "", true},              // don't use upperCamel
		{"title", lowerCamelCvt, args{"Normal"}, "", true},          // don't use upperCamel
		{"upperCamel", lowerCamelCvt, args{"CreateFail"}, "", true}, // don't use upperCamel

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.cvt.convertWithErr(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("lowerCamelConverter.convertWithErr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("lowerCamelConverter.convertWithErr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_lowerCamelConverter_unconvert(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name    string
		cvt     styleConverter
		args    args
		want    string
		wantErr bool
	}{
		{"lower", lowerCamelCvt, args{"success"}, "success", false},
		{"lowerCamel", lowerCamelCvt, args{"create_fail"}, "createFail", false},
		{"lowerCamelWithSpecial", lowerCamelCvt, args{"create_udb_fail"}, "createUdbFail", false}, // cannot reserve special word
		{"upper", lowerCamelCvt, args{"H_a"}, "", true},                                           // don't use upperCamel
		{"title", lowerCamelCvt, args{"Normal"}, "", true},                                        // don't use upperCamel
		{"upperCamel", lowerCamelCvt, args{"Create_fail"}, "", true},                              // don't use upperCamel

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.cvt.unconvertWithErr(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("lowerCamelConverter.unconvertWithErr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("lowerCamelConverter.unconvertWithErr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_upperCamelConverter_convert(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name    string
		cvt     styleConverter
		args    args
		want    string
		wantErr bool
	}{
		{"lower", upperCamelCvt, args{"Success"}, "success", false},
		{"lowerCamel", upperCamelCvt, args{"CreateFail"}, "create_fail", false},
		{"lowerCamelWithSpecial", upperCamelCvt, args{"CreateUDBFail"}, "create_udb_fail", false},
		{"upper", upperCamelCvt, args{"ha"}, "", true},              // don't use lowerCamel
		{"title", upperCamelCvt, args{"normal"}, "", true},          // don't use lowerCamel
		{"upperCamel", upperCamelCvt, args{"createFail"}, "", true}, // don't use lowerCamel
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.cvt.convertWithErr(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("upperCamelConverter.convertWithErr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("upperCamelConverter.convertWithErr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_upperCamelConverter_unconvert(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name    string
		cvt     styleConverter
		args    args
		want    string
		wantErr bool
	}{
		{"lower", upperCamelCvt, args{"success"}, "Success", false},
		{"lowerCamel", upperCamelCvt, args{"create_fail"}, "CreateFail", false},
		{"lowerCamelWithSpecial", upperCamelCvt, args{"create_udb_fail"}, "CreateUdbFail", false}, // cannot reserve special word
		{"upper", upperCamelCvt, args{"H_a"}, "", true},                                           // don't use upperCamel
		{"title", upperCamelCvt, args{"Normal"}, "", true},                                        // don't use upperCamel
		{"upperCamel", upperCamelCvt, args{"Create_fail"}, "", true},                              // don't use upperCamel
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.cvt.unconvertWithErr(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("upperCamelConverter.unconvertWithErr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("upperCamelConverter.unconvertWithErr() = %v, want %v", got, tt.want)
			}
		})
	}
}
