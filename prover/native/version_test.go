package native

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestBB6ToolchainSelection(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX test executable")
	}
	for _, version := range []string{"6.0.0", "5.2.1"} {
		t.Run(version, func(t *testing.T) {
			binary := filepath.Join(t.TempDir(), "bb")
			source := "#!/bin/sh\nprintf '%s\\n' '" + version + "'\n"
			if err := os.WriteFile(binary, []byte(source), 0700); err != nil {
				t.Fatal(err)
			}
			backend, err := NewNativeProver(NativeOptions{BBBinary: binary})
			if err != nil {
				t.Fatal(err)
			}
			err = backend.version(context.Background())
			if (err == nil) != (version == "6.0.0") {
				t.Fatalf("version %s: %v", version, err)
			}
		})
	}
}

func TestBB6NativeStableMsmEnvironment(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX test executable")
	}
	binary := filepath.Join(t.TempDir(), "bb")
	if err := os.WriteFile(binary, []byte("#!/bin/sh\nprintf '%s' \"$BB_MSM_LEGACY\"\n"), 0700); err != nil {
		t.Fatal(err)
	}
	previous, present := os.LookupEnv("BB_MSM_LEGACY")
	os.Unsetenv("BB_MSM_LEGACY")
	t.Cleanup(func() {
		if present {
			os.Setenv("BB_MSM_LEGACY", previous)
		} else {
			os.Unsetenv("BB_MSM_LEGACY")
		}
	})
	backend, err := NewNativeProver(NativeOptions{BBBinary: binary})
	if err != nil {
		t.Fatal(err)
	}
	output, err := backend.run(context.Background(), "--version")
	if err != nil || string(output) != "1" {
		t.Fatalf("default stable MSM: %q, %v", output, err)
	}
	t.Setenv("BB_MSM_LEGACY", "caller")
	output, err = backend.run(context.Background(), "--version")
	if err != nil || string(output) != "caller" {
		t.Fatalf("caller environment: %q, %v", output, err)
	}
}
