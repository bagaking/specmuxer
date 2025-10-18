package config

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	// ConfigDirName is the root folder under the workspace for SpecMuxer metadata.
	ConfigDirName = ".specmuxer"
	// ConfigFileName is the configuration document stored under ConfigDirName.
	ConfigFileName = "conf.yml"
	// StatsFileName is the aggregated statistics file stored under ConfigDirName.
	StatsFileName = "stats.yml"
	// SessionsDirName contains session YAML records.
	SessionsDirName = "sessions"
	// LogsDirName contains rotated log files.
	LogsDirName = "logs"

	dirPermission  = 0o700
	filePermission = 0o600

	defaultResumeMode         = ResumeModeManual
	defaultIdleThreshold      = 120
	defaultRedactionRuleLimit = 64
)

// ResumeMode controls whether SpecMuxer resumes sessions automatically.
type ResumeMode string

const (
	// ResumeModeManual resumes sessions only when explicitly requested.
	ResumeModeManual ResumeMode = "manual"
	// ResumeModeAutomatic resumes sessions automatically at startup.
	ResumeModeAutomatic ResumeMode = "automatic"
)

// Config captures persisted SpecMuxer configuration.
type Config struct {
	WorkspaceRoot        string                     `yaml:"workspace_root"`
	ProjectID            string                     `yaml:"project_id"`
	ResumeMode           ResumeMode                 `yaml:"resume_mode"`
	IdleThresholdSeconds int                        `yaml:"idle_threshold_seconds"`
	RedactionRules       []string                   `yaml:"redaction_rules,omitempty"`
	AdapterOverrides     map[string]AdapterOverride `yaml:"adapter_overrides,omitempty"`

	paths Paths `yaml:"-"`
}

// AdapterOverride captures per-adapter lifecycle overrides.
type AdapterOverride struct {
	Start        []string          `yaml:"start,omitempty"`
	Resume       []string          `yaml:"resume,omitempty"`
	Health       []string          `yaml:"health,omitempty"`
	Stop         []string          `yaml:"stop,omitempty"`
	ExtractState []string          `yaml:"extract_state,omitempty"`
	Env          map[string]string `yaml:"env,omitempty"`
	WorkingDir   string            `yaml:"working_dir,omitempty"`
	TimeoutSec   int               `yaml:"timeout_seconds,omitempty"`
}

// Paths exposes resolved configuration and data locations.
type Paths struct {
	ConfigDir   string
	ConfigPath  string
	StatsPath   string
	SessionsDir string
	LogsDir     string
}

// Load initializes workspace directories, enforces permissions, and returns configuration.
func Load(workspace string, overridePath string) (*Config, error) {
	if workspace == "" {
		return nil, errors.New("workspace path is required")
	}
	absWorkspace, err := filepath.Abs(workspace)
	if err != nil {
		return nil, fmt.Errorf("resolve workspace: %w", err)
	}

	configDir := filepath.Join(absWorkspace, ConfigDirName)
	if err := ensureDir(configDir, dirPermission); err != nil {
		return nil, err
	}

	sessionsDir := filepath.Join(configDir, SessionsDirName)
	if err := ensureDir(sessionsDir, dirPermission); err != nil {
		return nil, err
	}

	logsDir := filepath.Join(configDir, LogsDirName)
	if err := ensureDir(logsDir, dirPermission); err != nil {
		return nil, err
	}

	statsPath := filepath.Join(configDir, StatsFileName)
	if err := ensureFile(statsPath, filePermission); err != nil {
		return nil, err
	}

	configPath := overridePath
	if configPath == "" {
		configPath = filepath.Join(configDir, ConfigFileName)
	}

	cfg := defaultConfig(absWorkspace)
	cfg.paths = Paths{
		ConfigDir:   configDir,
		ConfigPath:  configPath,
		StatsPath:   statsPath,
		SessionsDir: sessionsDir,
		LogsDir:     logsDir,
	}

	if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		if err := writeConfig(configPath, cfg); err != nil {
			return nil, err
		}
		return cfg, nil
	} else if err != nil {
		return nil, fmt.Errorf("stat config: %w", err)
	}

	fileBytes, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	if err := yaml.Unmarshal(fileBytes, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	cfg.WorkspaceRoot = absWorkspace
	cfg.paths = Paths{
		ConfigDir:   configDir,
		ConfigPath:  configPath,
		StatsPath:   statsPath,
		SessionsDir: sessionsDir,
		LogsDir:     logsDir,
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Paths returns resolved runtime paths.
func (c *Config) Paths() Paths {
	return c.paths
}

func defaultConfig(workspace string) *Config {
	return &Config{
		WorkspaceRoot:        workspace,
		ProjectID:            projectHash(workspace),
		ResumeMode:           defaultResumeMode,
		IdleThresholdSeconds: defaultIdleThreshold,
		RedactionRules:       []string{},
		AdapterOverrides:     map[string]AdapterOverride{},
	}
}

func (c *Config) validate() error {
	if c.ProjectID == "" {
		c.ProjectID = projectHash(c.WorkspaceRoot)
	}
	if c.ResumeMode == "" {
		c.ResumeMode = defaultResumeMode
	}
	switch c.ResumeMode {
	case ResumeModeManual, ResumeModeAutomatic:
	default:
		return fmt.Errorf("invalid resume_mode: %s", c.ResumeMode)
	}

	if c.IdleThresholdSeconds <= 0 {
		c.IdleThresholdSeconds = defaultIdleThreshold
	}
	if len(c.RedactionRules) > defaultRedactionRuleLimit {
		return fmt.Errorf("too many redaction rules (%d > %d)", len(c.RedactionRules), defaultRedactionRuleLimit)
	}
	return nil
}

func ensureDir(path string, perm fs.FileMode) error {
	info, err := os.Stat(path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		if err := os.MkdirAll(path, perm); err != nil {
			return fmt.Errorf("create dir %q: %w", path, err)
		}
	case err != nil:
		return fmt.Errorf("stat dir %q: %w", path, err)
	default:
		if !info.IsDir() {
			return fmt.Errorf("%q exists but is not a directory", path)
		}
		if info.Mode().Perm() != perm {
			if err := os.Chmod(path, perm); err != nil {
				return fmt.Errorf("chmod %q to %o: %w", path, perm, err)
			}
		}
	}
	return nil
}

func ensureFile(path string, perm fs.FileMode) error {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		if err := os.WriteFile(path, []byte{}, perm); err != nil {
			return fmt.Errorf("create file %q: %w", path, err)
		}
		return nil
	} else if err != nil {
		return fmt.Errorf("stat file %q: %w", path, err)
	}
	if err := os.Chmod(path, perm); err != nil {
		return fmt.Errorf("chmod file %q: %w", path, err)
	}
	return nil
}

func writeConfig(path string, cfg *Config) error {
	content, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(path, content, filePermission); err != nil {
		return fmt.Errorf("write config %q: %w", path, err)
	}
	return nil
}

func projectHash(workspace string) string {
	sum := sha1.Sum([]byte(workspace))
	return hex.EncodeToString(sum[:8])
}
