package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/yankeguo/ezadmis"
	"github.com/yankeguo/rg"
	admission_v1 "k8s.io/api/admission/v1"
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
			URL     string         `json:"url"`
			Headers map[string]any `json:"headers"`
			Query   map[string]any `json:"query"`
			Form    map[string]any `json:"form"`
			JSON    map[string]any `json:"json"`
		} `json:"webhook"`
	} `json:"imageAutoMapping"`
	ImagePullSecrets []core_v1.LocalObjectReference `json:"imagePullSecrets"`
}

type ImageMapper struct {
	opts Options
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

func (opts Options) CreateMapper() (m *ImageMapper) {
	return &ImageMapper{opts: opts}
}

func (m *ImageMapper) Lookup(image string) (newImage string, ok bool) {
	image = standardizeImage(image)
	newImage, ok = m.opts.ImageMappings[image]
	return
}

func main() {
	debug, _ := strconv.ParseBool(os.Getenv("DEBUG"))

	server := ezadmis.NewWebhookServer(
		ezadmis.WebhookServerOptions{
			Debug: debug,
			Handler: func(ctx context.Context, req *admission_v1.AdmissionRequest, rw ezadmis.WebhookResponseWriter) (err error) {
				defer rg.Guard(&err)

				opts := rg.Must(loadOptions())

				im := opts.CreateMapper()

				buf := rg.Must(req.Object.MarshalJSON())

				var currentPod core_v1.Pod
				rg.Must0(json.Unmarshal(buf, &currentPod))

				var replaced bool

				for i, c := range currentPod.Spec.Containers {
					if newImage, ok := im.Lookup(c.Image); ok {
						rw.PatchReplace(fmt.Sprintf("/spec/containers/%d/image", i), newImage)
						replaced = true
					}
				}

				for i, c := range currentPod.Spec.InitContainers {
					if newImage, ok := im.Lookup(c.Image); ok {
						rw.PatchReplace(fmt.Sprintf("/spec/initContainers/%d/image", i), newImage)
						replaced = true
					}
				}

				if !replaced {
					return
				}

				if len(currentPod.Spec.ImagePullSecrets) == 0 {
					rw.PatchReplace("/spec/imagePullSecrets", opts.ImagePullSecrets)
				} else {
				next:
					for _, item := range opts.ImagePullSecrets {
						for _, currentItem := range currentPod.Spec.ImagePullSecrets {
							if currentItem.Name == item.Name {
								continue next
							}
						}
						rw.PatchAdd("/spec/imagePullSecrets/-", item)
					}
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
