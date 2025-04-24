package main

import (
	"encoding/json"
	"os"

	"github.com/yankeguo/rg"
	core_v1 "k8s.io/api/core/v1"
)

type Options struct {
	ImageMappings    map[string]string `json:"imageMappings"`
	ImageAutoMapping struct {
		Match struct {
			Namespaces []string `json:"namespaces"`
			Images     []string `json:"images"`
		} `json:"match"`
		Registry string `json:"registry"`
		Webhook  struct {
			Override struct {
				Registry string `json:"registry"`
			} `json:"override"`
			URL     string            `json:"url"`
			Headers map[string]string `json:"headers"`
			Query   map[string]string `json:"query"`
			Form    map[string]string `json:"form"`
			JSON    map[string]string `json:"json"`
		} `json:"webhook"`
	} `json:"imageAutoMapping"`
	ImagePullSecrets []core_v1.LocalObjectReference `json:"imagePullSecrets"`
}

func LoadOptions() (opts Options, err error) {
	defer rg.Guard(&err)
	buf := rg.Must(os.ReadFile("/config/replaceimage.json"))
	rg.Must0(json.Unmarshal(buf, &opts))

	// standardize image names
	mappings := make(map[string]string)
	for key, val := range opts.ImageMappings {
		mappings[standardizeImage(key)] = val
	}
	opts.ImageMappings = mappings
	return
}
