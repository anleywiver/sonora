// Package mediainfo extracts duration and basic tags from an audio file by
// shelling out to ffprobe (part of ffmpeg, installed in the worker image).
package mediainfo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type Info struct {
	DurationMs  int
	Title       string
	Artist      string
	Album       string
	TrackNumber *int
}

type probeFormat struct {
	Duration string            `json:"duration"`
	Tags     map[string]string `json:"tags"`
}

type probeOutput struct {
	Format probeFormat `json:"format"`
}

// Probe reads duration and tags (title/artist/album/track). Missing tags
// are left blank — the caller decides on fallbacks (e.g. "Unknown Artist").
func Probe(ctx context.Context, path string) (*Info, error) {
	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		path,
	)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("mediainfo: ffprobe: %w", err)
	}

	var parsed probeOutput
	if err := json.Unmarshal(out.Bytes(), &parsed); err != nil {
		return nil, fmt.Errorf("mediainfo: parse ffprobe output: %w", err)
	}

	durationSeconds, err := strconv.ParseFloat(parsed.Format.Duration, 64)
	if err != nil {
		return nil, fmt.Errorf("mediainfo: parse duration: %w", err)
	}

	tags := lowercaseKeys(parsed.Format.Tags)
	info := &Info{
		DurationMs: int(durationSeconds * 1000),
		Title:      tags["title"],
		Artist:     tags["artist"],
		Album:      tags["album"],
	}
	// ffprobe often reports "track" as "3/12" — take the numerator.
	if raw, ok := tags["track"]; ok {
		var n int
		if _, err := fmt.Sscanf(raw, "%d", &n); err == nil {
			info.TrackNumber = &n
		}
	}
	return info, nil
}

func lowercaseKeys(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[strings.ToLower(k)] = v
	}
	return out
}
