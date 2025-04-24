package main

import (
	"os"
	"strings"
)

func flattenImage(image string) string {
	comps := strings.Split(image, "/")
	if len(comps) < 2 {
		return image
	}
	var shim []string
	for _, item := range comps[:len(comps)-1] {
		// ignore duplicated path component
		if len(shim) > 0 && shim[len(shim)-1] == item {
			continue
		}
		shim = append(shim, item)
	}
	for i := range shim {
		shim[i] = strings.ReplaceAll(shim[i], ".", "-")
		shim[i] = strings.ReplaceAll(shim[i], "_", "-")
	}
	return strings.Join(shim, "-") + "-" + comps[len(comps)-1]
}

func standardizeImage(image string) string {
	if !strings.Contains(image, ":") {
		image += ":latest"
	}

	splits := strings.Split(image, "/")

	// single part
	if len(splits) < 2 {
		return "docker.io/library/" + image
	}

	// custom domain
	if strings.Contains(splits[0], ".") || strings.Contains(splits[0], ":") {
		// first part is docker.io and only 2 parts
		if len(splits) == 2 && splits[0] == "docker.io" {
			return "docker.io/library/" + splits[1]
		}
		return image
	}

	return "docker.io/" + image
}

func expandMap(src map[string]string, m map[string]string) map[string]string {
	out := make(map[string]string)
	for k, v := range src {
		v = os.Expand(v, func(name string) string {
			return m[name]
		})
		out[k] = v
	}
	return out
}
