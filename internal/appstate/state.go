package appstate

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

type State struct {
	LastSeenVersion   string    `json:"last_seen_version,omitempty"`
	LastUpdateCheckAt time.Time `json:"last_update_check_at,omitempty"`
}

func FilePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "Golt", "state.json"), nil
}

func Load() (State, string, error) {
	path, err := FilePath()
	if err != nil {
		return State{}, "", err
	}

	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return State{}, path, nil
		}
		return State{}, "", err
	}

	var st State
	if err := json.Unmarshal(b, &st); err != nil {
		return State{}, path, nil
	}

	return st, path, nil
}

func Save(path string, st State) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	tmp := path + ".tmp"
	b, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')

	if err := os.WriteFile(tmp, b, 0644); err != nil {
		return err
	}

	return os.Rename(tmp, path)
}
