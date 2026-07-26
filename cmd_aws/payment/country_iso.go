package main

import (
	"encoding/json"
	"os"
	"strings"
	"sync"
)

var (
	isoMap     map[string]string
	isoMapOnce sync.Once
)

func loadISOMap() map[string]string {
	isoMapOnce.Do(func() {
		isoMap = map[string]string{
			"malaysia": "my",
			"singapore": "sg",
			"indonesia": "id",
			"thailand": "th",
			"united states": "us",
			"united kingdom": "gb",
			"australia": "au",
		}
		b, err := os.ReadFile("countries_iso_slim.json")
		if err != nil {
			return
		}
		var m map[string]string
		if json.Unmarshal(b, &m) == nil {
			for k, v := range m {
				isoMap[strings.ToLower(k)] = strings.ToLower(v)
			}
		}
	})
	return isoMap
}

func countryNameToISO(name string) string {
	n := strings.ToLower(strings.TrimSpace(name))
	if n == "" {
		return ""
	}
	if len(n) == 2 {
		return n
	}
	if v, ok := loadISOMap()[n]; ok {
		return v
	}
	return n
}
