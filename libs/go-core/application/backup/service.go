// Package backup implements the Sprint 13 scheduled database backup: a
// pg_dump piped through gzip, uploaded to a Hetzner Storage Box over
// SFTP. See docs/decisions/0007-sprint13-observability-and-dr.md — CLI
// tools (pg_dump, sftp) are used deliberately instead of a pure-Go
// reimplementation.
package backup

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

var ErrNotConfigured = errors.New("backup: BACKUP_SSH_HOST is not set, skipping")

type Config struct {
	DatabaseURL string
	SSHHost     string
	SSHUser     string
	RemotePath  string
	SSHKeyPath  string
}

type Service struct {
	cfg    Config
	tmpDir string
}

func NewService(cfg Config, tmpDir string) *Service {
	return &Service{cfg: cfg, tmpDir: tmpDir}
}

// RunBackup dumps the database, gzips it, and uploads it over SFTP. The
// local dump file is always removed afterward, success or failure.
func (s *Service) RunBackup(ctx context.Context) error {
	if s.cfg.SSHHost == "" {
		return ErrNotConfigured
	}

	dumpPath, err := s.dumpDatabase(ctx)
	if err != nil {
		return fmt.Errorf("backup: dump database: %w", err)
	}
	defer os.Remove(dumpPath)

	if err := s.upload(ctx, dumpPath); err != nil {
		return fmt.Errorf("backup: upload: %w", err)
	}
	return nil
}

func (s *Service) dumpDatabase(ctx context.Context) (string, error) {
	if err := os.MkdirAll(s.tmpDir, 0o755); err != nil {
		return "", fmt.Errorf("prepare tmp dir: %w", err)
	}
	filename := fmt.Sprintf("sonora-backup-%s.sql.gz", time.Now().UTC().Format("20060102-150405"))
	dumpPath := filepath.Join(s.tmpDir, filename)

	out, err := os.Create(dumpPath)
	if err != nil {
		return "", fmt.Errorf("create dump file: %w", err)
	}
	defer out.Close()

	dump := exec.CommandContext(ctx, "pg_dump", s.cfg.DatabaseURL, "--format=plain", "--no-owner")
	gzipCmd := exec.CommandContext(ctx, "gzip")

	pipe, err := dump.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("pipe pg_dump: %w", err)
	}
	gzipCmd.Stdin = pipe
	gzipCmd.Stdout = out

	var dumpErr, gzipErr error
	if dumpErr = dump.Start(); dumpErr != nil {
		return "", fmt.Errorf("start pg_dump: %w", dumpErr)
	}
	if gzipErr = gzipCmd.Start(); gzipErr != nil {
		return "", fmt.Errorf("start gzip: %w", gzipErr)
	}
	dumpErr = dump.Wait()
	gzipErr = gzipCmd.Wait()
	if dumpErr != nil {
		return "", fmt.Errorf("pg_dump: %w", dumpErr)
	}
	if gzipErr != nil {
		return "", fmt.Errorf("gzip: %w", gzipErr)
	}
	return dumpPath, nil
}

func (s *Service) upload(ctx context.Context, localPath string) error {
	remoteTarget := fmt.Sprintf("%s@%s:%s/%s", s.cfg.SSHUser, s.cfg.SSHHost, s.cfg.RemotePath, filepath.Base(localPath))
	args := []string{"-i", s.cfg.SSHKeyPath, "-oStrictHostKeyChecking=accept-new", localPath, remoteTarget}
	cmd := exec.CommandContext(ctx, "scp", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("scp: %w (%s)", err, string(output))
	}
	return nil
}
