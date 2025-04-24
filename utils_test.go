package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStandardizeImage(t *testing.T) {
	require.Equal(t, "docker.io/library/nginx:latest", standardizeImage("nginx"))
	require.Equal(t, "docker.io/yankeguo/minit:latest", standardizeImage("yankeguo/minit"))
	require.Equal(t, "docker.io/library/alpine:3.18", standardizeImage("docker.io/alpine:3.18"))
	require.Equal(t, "gcr.io/hello:v2", standardizeImage("gcr.io/hello:v2"))
}

func TestFlattenImage(t *testing.T) {
	require.Equal(t, "docker-io-library-nginx:latest", flattenImage("docker.io/library/nginx:latest"))
	require.Equal(t, "docker-io-yankeguo-minit:latest", flattenImage("docker.io/yankeguo/minit:latest"))
	require.Equal(t, "docker-io-library-alpine:3.18", flattenImage("docker.io/library/alpine:3.18"))
	require.Equal(t, "gcr-io-hello:v2", flattenImage("gcr.io/hello:v2"))
}

func TestExpandMap(t *testing.T) {
	require.Equal(t, map[string]string{
		"field1": "value1",
		"field2": "hello:world",
	}, expandMap(map[string]string{
		"field1": "value1",
		"field2": "$SOURCE_IMAGE:$TARGET_IMAGE",
	}, map[string]string{
		"SOURCE_IMAGE": "hello",
		"TARGET_IMAGE": "world",
	}))
}
