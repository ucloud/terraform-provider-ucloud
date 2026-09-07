package unet

import "testing"

func TestStyleConvertersConvertWithErr(t *testing.T) {
	tests := []struct {
		name      string
		converter styleConverter
		input     string
		want      string
		wantErr   bool
	}{
		{name: "upper", converter: upperCvt, input: "LOCAL_SSD", want: "local_ssd"},
		{name: "upper no span", converter: upperCvt, input: "LOCALSSD", want: "localssd"},
		{name: "upper mixed", converter: upperCvt, input: "LoCal_ssd", wantErr: true},
		{name: "upper empty", converter: upperCvt},
		{name: "lower camel", converter: newLowerCamelConverter(nil), input: "createFail", want: "create_fail"},
		{name: "lower camel acronym", converter: newLowerCamelConverter(nil), input: "createUDBFail", want: "create_udb_fail"},
		{name: "lower camel acronym suffix", converter: newLowerCamelConverter(nil), input: "localSSD", want: "local_ssd"},
		{name: "lower camel uppercase", converter: newLowerCamelConverter(nil), input: "HA", wantErr: true},
		{name: "lower camel title", converter: newLowerCamelConverter(nil), input: "Normal", wantErr: true},
		{name: "lower camel upper camel", converter: newLowerCamelConverter(nil), input: "CreateFail", wantErr: true},
		{name: "upper camel single", converter: upperCamelCvt, input: "A", want: "a"},
		{name: "upper camel", converter: upperCamelCvt, input: "CreateFail", want: "create_fail"},
		{name: "upper camel acronym", converter: upperCamelCvt, input: "CreateUDBFail", want: "create_udb_fail"},
		{name: "upper camel lowercase", converter: upperCamelCvt, input: "ha", wantErr: true},
		{name: "upper camel lower camel", converter: upperCamelCvt, input: "createFail", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.converter.convertWithErr(test.input)
			if (err != nil) != test.wantErr {
				t.Fatalf("convertWithErr(%q) error = %v, wantErr %t", test.input, err, test.wantErr)
			}
			if got != test.want {
				t.Errorf("convertWithErr(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}

func TestStyleConvertersUnconvertWithErr(t *testing.T) {
	tests := []struct {
		name      string
		converter styleConverter
		input     string
		want      string
		wantErr   bool
	}{
		{name: "upper", converter: upperCvt, input: "local_ssd", want: "LOCAL_SSD"},
		{name: "upper no span", converter: upperCvt, input: "localssd", want: "LOCALSSD"},
		{name: "upper mixed", converter: upperCvt, input: "LoCal_SSD", wantErr: true},
		{name: "upper empty", converter: upperCvt},
		{name: "lower camel", converter: newLowerCamelConverter(nil), input: "create_fail", want: "createFail"},
		{name: "lower camel acronym", converter: newLowerCamelConverter(nil), input: "create_udb_fail", want: "createUdbFail"},
		{name: "lower camel single", converter: newLowerCamelConverter(nil), input: "a", want: "a"},
		{name: "lower camel uppercase", converter: newLowerCamelConverter(nil), input: "H_a", wantErr: true},
		{name: "lower camel title", converter: newLowerCamelConverter(nil), input: "Normal", wantErr: true},
		{name: "lower camel upper camel", converter: newLowerCamelConverter(nil), input: "Create_fail", wantErr: true},
		{name: "upper camel", converter: upperCamelCvt, input: "create_fail", want: "CreateFail"},
		{name: "upper camel acronym", converter: upperCamelCvt, input: "create_udb_fail", want: "CreateUdbFail"},
		{name: "upper camel single", converter: upperCamelCvt, input: "a", want: "A"},
		{name: "upper camel uppercase", converter: upperCamelCvt, input: "H_a", wantErr: true},
		{name: "upper camel title", converter: upperCamelCvt, input: "Normal", wantErr: true},
		{name: "upper camel upper camel", converter: upperCamelCvt, input: "Create_fail", wantErr: true},
		{name: "upper camel empty", converter: upperCamelCvt},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.converter.unconvertWithErr(test.input)
			if (err != nil) != test.wantErr {
				t.Fatalf("unconvertWithErr(%q) error = %v, wantErr %t", test.input, err, test.wantErr)
			}
			if got != test.want {
				t.Errorf("unconvertWithErr(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}

func TestStyleConvertersRoundTrip(t *testing.T) {
	tests := []struct {
		name          string
		converter     styleConverter
		input         string
		canonical     string
		reconstructed string
	}{
		{name: "upper", converter: upperCvt, input: "LOCAL_SSD", canonical: "local_ssd", reconstructed: "LOCAL_SSD"},
		{name: "lower camel", converter: newLowerCamelConverter(nil), input: "createFail", canonical: "create_fail", reconstructed: "createFail"},
		{name: "lower camel acronym", converter: newLowerCamelConverter(nil), input: "createUDBFail", canonical: "create_udb_fail", reconstructed: "createUdbFail"},
		{name: "upper camel", converter: upperCamelCvt, input: "CreateFail", canonical: "create_fail", reconstructed: "CreateFail"},
		{name: "upper camel acronym", converter: upperCamelCvt, input: "CreateUDBFail", canonical: "create_udb_fail", reconstructed: "CreateUdbFail"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			converted, err := test.converter.convertWithErr(test.input)
			if err != nil {
				t.Fatalf("convertWithErr(%q): %v", test.input, err)
			}
			if converted != test.canonical {
				t.Fatalf("convertWithErr(%q) = %q, want %q", test.input, converted, test.canonical)
			}

			reconstructed, err := test.converter.unconvertWithErr(converted)
			if err != nil {
				t.Fatalf("unconvertWithErr(%q): %v", converted, err)
			}
			if reconstructed != test.reconstructed {
				t.Fatalf("unconvertWithErr(%q) = %q, want %q", converted, reconstructed, test.reconstructed)
			}

			reconverted, err := test.converter.convertWithErr(reconstructed)
			if err != nil {
				t.Fatalf("convertWithErr(%q): %v", reconstructed, err)
			}
			if reconverted != test.canonical {
				t.Errorf("convertWithErr(%q) = %q, want canonical %q", reconstructed, reconverted, test.canonical)
			}
		})
	}
}
