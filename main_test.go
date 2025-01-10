package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStandardizeImage(t *testing.T) {
	require.Equal(t, "docker.io/library/nginx:latest", standardizeImage("nginx"))
	require.Equal(t, "docker.io/yankeguo/minit:latest", standardizeImage("yankeguo/minit"))
	require.Equal(t, "gcr.io/hello:v2", standardizeImage("gcr.io/hello:v2"))
}
