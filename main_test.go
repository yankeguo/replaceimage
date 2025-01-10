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
