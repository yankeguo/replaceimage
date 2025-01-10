package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/yankeguo/ezadmis"
	"github.com/yankeguo/rg"
	admission_v1 "k8s.io/api/admission/v1"
	core_v1 "k8s.io/api/core/v1"
)

type Options struct {
	ImageMappings    map[string]string              `json:"imageMappings"`
	ImagePullSecrets []core_v1.LocalObjectReference `json:"imagePullSecrets"`
}

func loadOptions() (opts Options, err error) {
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

func standardizeImage(image string) string {
	if !strings.Contains(image, ":") {
		image += ":latest"
	}

	splits := strings.Split(image, "/")

	if len(splits) > 1 {
		if strings.Contains(splits[0], ".") || strings.Contains(splits[0], ":") {
			return image
		} else {
			return "docker.io/" + image
		}
	}

	return "docker.io/library/" + image
}

func main() {
	server := ezadmis.NewWebhookServer(
		ezadmis.WebhookServerOptions{
			Handler: func(ctx context.Context, req *admission_v1.AdmissionRequest, rw ezadmis.WebhookResponseWriter) (err error) {
				defer rg.Guard(&err)

				opts := rg.Must(loadOptions())

				buf := rg.Must(req.Object.MarshalJSON())

				var currentPod core_v1.Pod
				rg.Must0(json.Unmarshal(buf, &currentPod))

				for i, c := range currentPod.Spec.Containers {
					if newImage, ok := opts.ImageMappings[standardizeImage(c.Image)]; ok {
						rw.PatchReplace(fmt.Sprintf("/spec/containers/%d/image", i), newImage)
					}
				}

				for i, c := range currentPod.Spec.InitContainers {
					if newImage, ok := opts.ImageMappings[standardizeImage(c.Image)]; ok {
						rw.PatchReplace(fmt.Sprintf("/spec/initContainers/%d/image", i), newImage)
					}
				}

			next:
				for _, item := range opts.ImagePullSecrets {
					for _, currentItem := range currentPod.Spec.ImagePullSecrets {
						if currentItem.Name == item.Name {
							continue next
						}
					}
					rw.PatchAdd("/spec/imagePullSecrets/-", item)
				}

				return
			},
		},
	)

	chErr := make(chan error, 1)
	chSig := make(chan os.Signal, 1)
	signal.Notify(chSig, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		chErr <- server.ListenAndServe()
	}()

	select {
	case err := <-chErr:
		if err == nil {
			return
		}
		log.Println("error:", err.Error())
		os.Exit(1)
	case sig := <-chSig:
		log.Println("signal:", sig.String())
	}

	server.Shutdown(context.Background())
}
