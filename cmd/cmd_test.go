package cmd_test

import (
	"bytes"
	"os"
	"pdn/cmd"
	"pdn/signal"
	"testing"
)

// ParseArgs parses the command-line arguments and returns the configuration.
// It returns an error if the arguments are invalid.
func TestParseArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    signal.Config
		wantErr bool
	}{
		{
			name: "given valid args when parsed then return config",
			args: []string{"-port=8080", "-key=/path/to/key.pem", "-cert=/path/to/cert.pem"},
			want: signal.Config{Port: 8080, KeyFile: "/path/to/key.pem", CertFile: "/path/to/cert.pem"},
		},
		{
			name: "given missing port when parsed then return config with default port",
			args: []string{"-key=/path/to/key.pem", "-cert=/path/to/cert.pem"},
			want: signal.Config{Port: signal.DefaultPort, KeyFile: "/path/to/key.pem", CertFile: "/path/to/cert.pem"},
		},
		{
			name:    "given empty key file when parsed then return config with empty key file",
			args:    []string{"-port=8080", "-cert=/path/to/cert.pem"},
			want:    signal.Config{Port: 8080, KeyFile: "", CertFile: "/path/to/cert.pem"},
			wantErr: false,
		},
		{
			name:    "given empty cert file when parsed then return config with empty cert file",
			args:    []string{"-port=8080", "-key=/path/to/key.pem"},
			want:    signal.Config{Port: 8080, KeyFile: "/path/to/key.pem", CertFile: ""},
			wantErr: false,
		},
		{
			name: "given no args when parsed then return config",
			args: []string{},
			want: signal.Config{Port: signal.DefaultPort, KeyFile: "", CertFile: ""},
		},
		{
			name:    "given extra args when parsed then return error",
			args:    []string{"-port=8080", "-key=/path/to/key.pem", "-cert=/path/to/cert.pem", "extra"},
			wantErr: true,
		},
		{
			name:    "given invalid flag format when parsed then return error",
			args:    []string{"-extra"},
			wantErr: true,
		},
		{
			name:    "given invalid non-flag args when parsed then return error",
			args:    []string{"port"},
			wantErr: true,
		},
		{
			name:    "given port flag without value when parsed then return error",
			args:    []string{"-port"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			got, err := cmd.ParseArgs(&output, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !got.IsSame(tt.want) {
				t.Errorf("ParseArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function to create a temporary file and return its path
func createTempFile() (string, error) {
	tmpFile, err := os.CreateTemp("", "testfile")
	if err != nil {
		return "", err
	}
	if closeErr := tmpFile.Close(); closeErr != nil {
		return "", closeErr
	}
	return tmpFile.Name(), nil
}

// TestSetupConfig tests the SetupConfig function, including handling errors from ParseArgs and Config.Validate.
func TestSetupConfig(t *testing.T) {
	keyFile, err := createTempFile()
	if err != nil {
		t.Fatalf("Failed to create temporary key file: %v", err)
	}
	certFile, err := createTempFile()
	if err != nil {
		t.Fatalf("Failed to create temporary cert file: %v", err)
	}

	// Clean up temporary files after the test
	defer func() {
		_ = os.Remove(keyFile)
		_ = os.Remove(certFile)
	}()

	tests := []struct {
		name          string
		args          []string
		expected      signal.Config
		parseError    bool
		validateError bool
	}{
		{
			name: "given valid args when setup config then return valid config",
			args: []string{"-port=8080", "-key=" + keyFile, "-cert=" + certFile},
			expected: signal.Config{
				Port:     8080,
				KeyFile:  keyFile,
				CertFile: certFile,
			},
			parseError:    false,
			validateError: false,
		},
		{
			name: "given no args when setup config then return default config",
			args: []string{},
			expected: signal.Config{
				Port:     signal.DefaultPort,
				KeyFile:  "",
				CertFile: "",
			},
			parseError:    false,
			validateError: false,
		},
		{
			name:          "given invalid port value when setup config then return error",
			args:          []string{"-port=70000"},
			parseError:    false,
			validateError: true,
		},
		{
			name:          "given non-existent cert file when setup config then return error",
			args:          []string{"-port=8080", "-key=" + keyFile, "-cert=/non/existent/cert.pem"},
			parseError:    false,
			validateError: true,
		},
		{
			name:          "given non-existent key file when setup config then return error",
			args:          []string{"-port=8080", "-cert=" + certFile, "-key=/non/existent/key.pem"},
			parseError:    false,
			validateError: true,
		},
		{
			name:          "given invalid flag format when setup config then return error",
			args:          []string{"-extra"},
			parseError:    true,
			validateError: false, // No need to check Validate if ParseArgs fails
		},
		{
			name:          "given port flag without value when setup config then return error",
			args:          []string{"-port"},
			parseError:    true,
			validateError: false, // No need to check Validate if ParseArgs fails
		},
		{
			name: "given empty key file and cert file when setup config then return valid config",
			args: []string{"-port=8080"},
			expected: signal.Config{
				Port:     8080,
				KeyFile:  "",
				CertFile: "",
			},
			parseError:    false,
			validateError: false,
		},
		{
			name:          "given empty key file and non-empty cert file when setup config then return error",
			args:          []string{"-port=8080", "-cert=" + certFile},
			parseError:    false,
			validateError: true,
		},
		{
			name:          "given non-empty key file and empty cert file when setup config then return error",
			args:          []string{"-port=8080", "-key=" + keyFile},
			parseError:    false,
			validateError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			originalArgs := os.Args
			defer func() { os.Args = originalArgs }()

			os.Args = append([]string{"cmd"}, tt.args...)

			config, err := cmd.SetupConfig()

			if tt.parseError {
				if err == nil {
					t.Errorf("SetupConfig() expected ParseArgs error, got nil")
				}
				return
			}

			if tt.validateError {
				if err == nil {
					t.Errorf("SetupConfig() expected Validate error, got nil")
				}
				return
			}

			if !config.IsSame(tt.expected) {
				t.Errorf("%s: SetupConfig() = %v, expected %v", tt.name, config, tt.expected)
			}
		})
	}
}
