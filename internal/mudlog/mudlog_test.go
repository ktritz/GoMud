package mudlog

import "testing"

func TestErrorBeforeSetupLoggerDoesNotPanic(t *testing.T) {
	slogInstance = nil

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Error() panicked before SetupLogger(): %v", r)
		}
	}()

	Error("pre-setup log", "case", "test")
}
