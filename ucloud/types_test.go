package ucloud

import (
	"reflect"
	"testing"
)

func Test_parseInstanceType(t *testing.T) {
	type args struct {
		s string
	}

	tests := []struct {
		name    string
		args    args
		want    *instanceType
		wantErr bool
	}{
		{"ok_highcpu", args{"n-highcpu-1"}, &instanceType{1, 1024, "n", "highcpu"}, false},
		{"ok_basic", args{"n-basic-1"}, &instanceType{1, 2048, "n", "basic"}, false},
		{"ok_standard", args{"n-standard-1"}, &instanceType{1, 4096, "n", "standard"}, false},
		{"ok_highmem", args{"n-highmem-1"}, &instanceType{1, 8192, "n", "highmem"}, false},
		{"ok_customized", args{"n-customized-1-12"}, &instanceType{1, 12288, "n", "customized"}, false},
		{"ok_customized_ratio", args{"n-customized-8-12"}, &instanceType{8, 12288, "n", "customized"}, false},
		{"ok_customized_ratio_opposite", args{"n-customized-12-8"}, &instanceType{12, 8192, "n", "customized"}, false},

		{"err_customized_ratio_opposite", args{"n-customized-14-6"}, nil, true},
		{"err_customized_core", args{"n-customized-3-11"}, nil, true},
		{"err_customized_ratio", args{"n-customized-4-50"}, nil, true},
		{"err_customized", args{"n-customized-1-5"}, nil, true},
		{"err_type", args{"nx-highcpu-1"}, nil, true},
		{"err_scale_type", args{"n-invalid-1"}, nil, true},
		{"err_cpu_too_much", args{"n-highcpu-33"}, nil, true},
		{"err_cpu_too_less", args{"n-highcpu-0"}, nil, true},
		{"err_cpu_is_invalid", args{"n-highcpu-x"}, nil, true},
		{"err_customized_format_len", args{"n-customized-1"}, nil, true},
		{"err_customized_format_number", args{"n-customized-x"}, nil, true},
		{"err_customized_should_be_standard", args{"n-customized-1-2"}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseInstanceType(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseInstanceType() arg %s got %#v error = %v, wantErr %v", tt.args.s, got, err, tt.wantErr)
				return
			}

			if got == nil {
				return
			}

			if !(tt.want.CPU == got.CPU) ||
				!(tt.want.Memory == got.Memory) ||
				!(tt.want.HostType == got.HostType) ||
				!(tt.want.HostScaleType == got.HostScaleType) {
				t.Errorf("parseInstanceType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseAssociationInfo(t *testing.T) {
	type args struct {
		assocId string
	}
	tests := []struct {
		name    string
		args    args
		want    *associationInfo
		wantErr bool
	}{
		{
			"ok",
			args{"eip#eip-xxx:uhost#uhost-xxx"},
			&associationInfo{"eip", "eip-xxx", "uhost", "uhost-xxx"},
			false,
		},
		{"err_no_colon", args{"eip#eip-xxx-uhost#uhost-xxx"}, nil, true},
		{"err_empty", args{""}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseAssociationInfo(tt.args.assocId)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseAssociationInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseAssociationInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseUCloudCidrBlock(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    *cidrBlock
		wantErr bool
	}{
		{"ok_10.0", args{"10.0.0.0/8"}, &cidrBlock{"10.0.0.0", 8}, false},
		{"err_10.0", args{"10.1.0.0/8"}, nil, true},
		{"ok_10.128", args{"10.128.0.0/9"}, &cidrBlock{"10.128.0.0", 9}, false},
		{"err_10.128", args{"10.128.0.0/8"}, nil, true},
		{"ok_10.10", args{"10.10.10.248/29"}, &cidrBlock{"10.10.10.248", 29}, false},
		{"err_10.10", args{"10.10.10.249/29"}, nil, true},
		{"ok_172.16", args{"172.16.0.0/12"}, &cidrBlock{"172.16.0.0", 12}, false},
		{"err_172.16", args{"172.16.1.0/12"}, nil, true},
		{"ok_172.31", args{"172.31.128.0/17"}, &cidrBlock{"172.31.128.0", 17}, false},
		{"err_172.31", args{"172.31.1.0/17"}, nil, true},
		{"ok_172.18", args{"172.18.255.248/29"}, &cidrBlock{"172.18.255.248", 29}, false},
		{"err_172.18", args{"172.18.255.6/29"}, nil, true},
		{"ok_192.0", args{"192.168.0.0/16"}, &cidrBlock{"192.168.0.0", 16}, false},
		{"err_192.0", args{"192.168.0.1/16"}, nil, true},
		{"ok_192.128", args{"192.168.128.0/17"}, &cidrBlock{"192.168.128.0", 17}, false},
		{"err_192.128", args{"192.168.128.1/17"}, nil, true},
		{"ok_192.255", args{"192.168.255.248/29"}, &cidrBlock{"192.168.255.248", 29}, false},
		{"err_192.255", args{"192.168.255.1/17"}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseUCloudCidrBlock(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseUCloudCidrBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseUCloudCidrBlock() got = %v, want %v", got, tt.want)
			}
		})
	}
}
