// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"encoding/json"
	"os"
	"sync"
)

type presetStore struct {
	mu   sync.Mutex
	path string
}

var globalPresetStore *presetStore

func initPresetStore(path string) {
	globalPresetStore = &presetStore{path: path}
}

func (s *presetStore) load() ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}

	var presets []string
	if err := json.Unmarshal(data, &presets); err != nil {
		return nil, err
	}
	return presets, nil
}

func (s *presetStore) add(keyword string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var presets []string
	data, err := os.ReadFile(s.path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if len(data) > 0 {
		if err := json.Unmarshal(data, &presets); err != nil {
			return err
		}
	}

	for _, p := range presets {
		if p == keyword {
			return nil
		}
	}
	presets = append(presets, keyword)
	return s.writeFile(presets)
}

func (s *presetStore) remove(keyword string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	var presets []string
	if err := json.Unmarshal(data, &presets); err != nil {
		return err
	}

	filtered := presets[:0]
	for _, p := range presets {
		if p != keyword {
			filtered = append(filtered, p)
		}
	}
	return s.writeFile(filtered)
}

func (s *presetStore) writeFile(presets []string) error {
	data, err := json.MarshalIndent(presets, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}
