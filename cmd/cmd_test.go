package cmd_test

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"os"
	"pdn/cmd"
	"pdn/signal"
	"testing"
)

// parse parses the command-line arguments and returns the configuration.
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
			got, err := cmd.Parse(&output, tt.args)
			if tt.wantErr {
				assert.Errorf(t, err, "parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
			assert.Truef(t, got.Signal.IsSame(tt.want), "parse() = %v, want %v", got, tt.want)
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

// TestSetupConfig tests the SetupConfig function, including handling errors from parse and Config.Validate.
func TestSetupConfig(t *testing.T) {
	keyFile, err := createTempFile()
	if err != nil {
		assert.Errorf(t, err, "Failed to create temporary key file: %v", err)
		return
	}
	certFile, err := createTempFile()
	if err != nil {
		assert.Errorf(t, err, "Failed to create temporary cert file: %v", err)
		return
	}

	// Clean up temporary files after the test
	defer func() {
		_ = os.Remove(keyFile)
		_ = os.Remove(certFile)
	}()

	tests := []struct {
		name                string
		args                []string
		expected            signal.Config
		expectParseError    bool
		expectValidateError bool
	}{
		{
			name: "given valid args when setup config then return valid config",
			args: []string{"-port=8080", "-key=" + keyFile, "-cert=" + certFile},
			expected: signal.Config{
				Port:     8080,
				KeyFile:  keyFile,
				CertFile: certFile,
			},
			expectParseError:    false,
			expectValidateError: false,
		},
		{
			name: "given no args when setup config then return default config",
			args: []string{},
			expected: signal.Config{
				Port:     signal.DefaultPort,
				KeyFile:  "",
				CertFile: "",
			},
			expectParseError:    false,
			expectValidateError: false,
		},
		{
			name:                "given invalid port value when setup config then return error",
			args:                []string{"-port=70000"},
			expectParseError:    false,
			expectValidateError: true,
		},
		{
			name:                "given non-existent cert file when setup config then return error",
			args:                []string{"-port=8080", "-key=" + keyFile, "-cert=/non/existent/cert.pem"},
			expectParseError:    false,
			expectValidateError: true,
		},
		{
			name:                "given non-existent key file when setup config then return error",
			args:                []string{"-port=8080", "-cert=" + certFile, "-key=/non/existent/key.pem"},
			expectParseError:    false,
			expectValidateError: true,
		},
		{
			name:                "given invalid flag format when setup config then return error",
			args:                []string{"-extra"},
			expectParseError:    true,
			expectValidateError: false, // No need to check Validate if parse fails
		},
		{
			name:                "given port flag without value when setup config then return error",
			args:                []string{"-port"},
			expectParseError:    true,
			expectValidateError: false, // No need to check Validate if parse fails
		},
		{
			name: "given empty key file and cert file when setup config then return valid config",
			args: []string{"-port=8080"},
			expected: signal.Config{
				Port:     8080,
				KeyFile:  "",
				CertFile: "",
			},
			expectParseError:    false,
			expectValidateError: false,
		},
		{
			name:                "given empty key file and non-empty cert file when setup config then return error",
			args:                []string{"-port=8080", "-cert=" + certFile},
			expectParseError:    false,
			expectValidateError: true,
		},
		{
			name:                "given non-empty key file and empty cert file when setup config then return error",
			args:                []string{"-port=8080", "-key=" + keyFile},
			expectParseError:    false,
			expectValidateError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			buf := bytes.NewBuffer(make([]byte, 1024))

			config, err := cmd.SetupConfig(buf, tt.args)

			if tt.expectParseError {
				assert.Error(t, err)
				return
			}

			if tt.expectValidateError {
				assert.Error(t, err)
				return
			}

			if !config.Signal.IsSame(tt.expected) {
				assert.Errorf(t, err, "SetupConfig() = %v, expected %v", config, tt.expected)
			}
		})
	}
}
