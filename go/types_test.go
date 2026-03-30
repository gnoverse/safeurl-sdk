package safeurl

import (
	"testing"
)

func TestScanState_IsTerminal(t *testing.T) {
	tests := []struct {
		state    ScanState
		terminal bool
	}{
		{ScanStatePending, false},
		{ScanStateProcessing, false},
		{ScanStateCompleted, true},
		{ScanStateFailed, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			if got := tt.state.IsTerminal(); got != tt.terminal {
				t.Errorf("ScanState(%q).IsTerminal() = %v, want %v", tt.state, got, tt.terminal)
			}
		})
	}
}

func TestVerdict_IsSafe(t *testing.T) {
	tests := []struct {
		verdict Verdict
		safe    bool
	}{
		{VerdictSafe, true},
		{VerdictMalicious, false},
		{VerdictSuspect, false},
		{VerdictUnknown, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.verdict), func(t *testing.T) {
			if got := tt.verdict.IsSafe(); got != tt.safe {
				t.Errorf("Verdict(%q).IsSafe() = %v, want %v", tt.verdict, got, tt.safe)
			}
		})
	}
}

func TestVerdict_IsUnsafe(t *testing.T) {
	tests := []struct {
		verdict Verdict
		unsafe  bool
	}{
		{VerdictSafe, false},
		{VerdictMalicious, true},
		{VerdictSuspect, true},
		{VerdictUnknown, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.verdict), func(t *testing.T) {
			if got := tt.verdict.IsUnsafe(); got != tt.unsafe {
				t.Errorf("Verdict(%q).IsUnsafe() = %v, want %v", tt.verdict, got, tt.unsafe)
			}
		})
	}
}

func TestScanResponse_IsComplete(t *testing.T) {
	tests := []struct {
		name     string
		state    ScanState
		complete bool
	}{
		{"pending", ScanStatePending, false},
		{"processing", ScanStateProcessing, false},
		{"completed", ScanStateCompleted, true},
		{"failed", ScanStateFailed, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ScanResponse{State: tt.state}
			if got := s.IsComplete(); got != tt.complete {
				t.Errorf("ScanResponse{State: %q}.IsComplete() = %v, want %v", tt.state, got, tt.complete)
			}
		})
	}
}
