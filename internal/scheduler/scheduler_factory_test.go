package scheduler

import (
	"testing"
)

// TestSchedulerFactory verifies scheduler creation and mode selection.
func TestSchedulerFactory_ModeSelection(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "timingwheel mode",
			cfg: &Config{
				Mode: ModeTimingWheel,
			},
			wantErr: false,
		},
		{
			name: "asynq mode with redis",
			cfg: &Config{
				Mode: ModeAsynq,
				Asynq: AsynqConfig{
					RedisURL: "redis://localhost:6379",
				},
			},
			wantErr: false,
		},
		{
			name: "asynq mode without redis",
			cfg: &Config{
				Mode: ModeAsynq,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := New(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && s == nil {
				t.Error("New() returned nil scheduler with no error")
			}
		})
	}
}

// TestDetectMode verifies mode detection logic.
func TestDetectMode(t *testing.T) {
	tests := []struct {
		name         string
		explicitMode string
		dbDriver     string
		want         ScheduleMode
	}{
		{
			name:         "explicit timingwheel",
			explicitMode: string(ModeTimingWheel),
			dbDriver:     "sqlite",
			want:         ModeTimingWheel,
		},
		{
			name:         "explicit asynq",
			explicitMode: string(ModeAsynq),
			dbDriver:     "sqlite",
			want:         ModeAsynq,
		},
		{
			name:         "postgres defaults to asynq",
			explicitMode: "",
			dbDriver:     "postgres",
			want:         ModeAsynq,
		},
		{
			name:         "sqlite defaults to timingwheel",
			explicitMode: "",
			dbDriver:     "sqlite",
			want:         ModeTimingWheel,
		},
		{
			name:         "empty driver defaults to timingwheel",
			explicitMode: "",
			dbDriver:     "",
			want:         ModeTimingWheel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectMode(tt.explicitMode, tt.dbDriver)
			if got != tt.want {
				t.Errorf("DetectMode() = %v, want %v", got, tt.want)
			}
		})
	}
}
