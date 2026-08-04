package mediainfo

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

const waveformPeakCount = 200

// GenerateWaveform decodes path to raw 8-bit unsigned mono PCM at 8kHz via
// ffmpeg, then reduces it to peakCount buckets (peak absolute deviation
// from the 128 midpoint, 0-255) — a compact enough shape for a UI
// waveform without pulling in a dedicated audio-decoding library.
func GenerateWaveform(ctx context.Context, path string) ([]int16, error) {
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-v", "quiet",
		"-i", path,
		"-ac", "1",
		"-filter:a", "aresample=8000",
		"-f", "u8",
		"pipe:1",
	)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("mediainfo: ffmpeg waveform decode: %w", err)
	}

	samples := out.Bytes()
	if len(samples) == 0 {
		return nil, fmt.Errorf("mediainfo: ffmpeg produced no samples")
	}

	bucketCount := waveformPeakCount
	if len(samples) < bucketCount {
		bucketCount = len(samples)
	}
	bucketSize := len(samples) / bucketCount

	peaks := make([]int16, bucketCount)
	for i := 0; i < bucketCount; i++ {
		start := i * bucketSize
		end := start + bucketSize
		if i == bucketCount-1 {
			end = len(samples)
		}
		var peak int16
		for _, sample := range samples[start:end] {
			deviation := int16(sample) - 128
			if deviation < 0 {
				deviation = -deviation
			}
			if deviation > peak {
				peak = deviation
			}
		}
		peaks[i] = peak
	}
	return peaks, nil
}
