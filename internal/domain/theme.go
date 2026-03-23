package domain

import "encoding/json"

type ThemeFile struct {
	Schema string                         `json:"$schema,omitempty"`
	Name   string                         `json:"name"`
	Defs   map[string]string              `json:"defs"`
	Theme  map[string]json.RawMessage     `json:"theme"`
}

type ThemeToken struct {
	Dark  string `json:"dark"`
	Light string `json:"light"`
}
