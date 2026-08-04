export interface LyricLine {
  timeMs: number;
  text: string;
}

// Parses standard LRC-format synced lyrics: lines like "[01:23.45] text".
// Lines without a timestamp (rare, sometimes metadata) are dropped since
// there's nothing to sync them to.
export function parseLRC(synced: string): LyricLine[] {
  const lines: LyricLine[] = [];
  for (const raw of synced.split("\n")) {
    const match = raw.match(/^\[(\d{2}):(\d{2})\.(\d{2,3})\]\s*(.*)$/);
    if (!match) continue;
    const [, mm, ss, frac, text] = match;
    const fracMs = frac.length === 2 ? Number(frac) * 10 : Number(frac);
    const timeMs = Number(mm) * 60_000 + Number(ss) * 1000 + fracMs;
    lines.push({ timeMs, text: text.trim() });
  }
  return lines;
}

// Index of the last line whose timestamp has passed — the "active" line.
export function activeLineIndex(lines: LyricLine[], positionMs: number): number {
  let idx = -1;
  for (let i = 0; i < lines.length; i++) {
    if (lines[i].timeMs <= positionMs) idx = i;
    else break;
  }
  return idx;
}
