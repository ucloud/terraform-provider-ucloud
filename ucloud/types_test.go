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
		{"ok_customized", args{"n-customized-1-5"}, &instanceType{1, 5120, "n", "customized"}, false},

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
		{"ok", args{"192.168.1.0/24"}, &cidrBlock{"192.168.1.0", 24}, false},

		{"err_ip", args{"1.0.0.0/24"}, nil, true},
		{"err_ip_conflict_with_mask", args{"192.168.1.1/24"}, nil, true},

		{"err_ip_range_192_network", args{"192.167.1.0/24"}, nil, true},
		{"err_ip_range_192_mask_too_small", args{"192.168.1.0/15"}, nil, true},
		{"err_ip_range_192_mask_too_large", args{"192.168.1.0/30"}, nil, true},

		{"err_ip_range_172_network_too_small", args{"172.15.1.0/24"}, nil, true},
		{"err_ip_range_172_network_too_large", args{"172.32.1.0/24"}, nil, true},
		{"err_ip_range_172_mask_too_small", args{"172.16.1.0/11"}, nil, true},
		{"err_ip_range_172_mask_too_large", args{"172.16.1.0/30"}, nil, true},

		{"err_ip_range_10_mask_too_small", args{"10.0.1.0/7"}, nil, true},
		{"err_ip_range_10_mask_too_large", args{"10.0.1.0/30"}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseUCloudCidrBlock(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseUClounillock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseUCloudCidrBlock() = %v, want %v", got, tt.want)
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
