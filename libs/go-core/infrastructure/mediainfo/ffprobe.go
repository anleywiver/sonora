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
	// Genre and Year back the Sprint 14 ingest filter rules (ADR 0008) —
	// both are "" / nil when the file has no such tag, which the filter
	// treats as "doesn't apply" rather than a rejection.
	Genre string
	Year  *int
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
		Genre:      tags["genre"],
	}
	// ffprobe often reports "track" as "3/12" — take the numerator.
	if raw, ok := tags["track"]; ok {
		var n int
		if _, err := fmt.Sscanf(raw, "%d", &n); err == nil {
			info.TrackNumber = &n
		}
	}
	// "date" is the more common ID3v2 tag; some files use "year" instead.
	// Either can be a bare year or a full date ("2023" or "2023-05-01") —
	// Sscanf %d just reads the leading digits either way.
	if raw, ok := tags["date"]; ok {
		if year, ok := parseYear(raw); ok {
			info.Year = &year
		}
	} else if raw, ok := tags["year"]; ok {
		if year, ok := parseYear(raw); ok {
			info.Year = &year
		}
	}
	return info, nil
}

func parseYear(raw string) (int, bool) {
	var year int
	if _, err := fmt.Sscanf(raw, "%d", &year); err != nil || year < 1000 || year > 9999 {
		return 0, false
	}
	return year, true
}

func lowercaseKeys(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[strings.ToLower(k)] = v
	}
	return out
}
