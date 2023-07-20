package dockerutil

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type PullEvent struct {
	Status         string `json:"status"`
	Error          string `json:"error"`
	Progress       string `json:"progress"`
	ProgressDetail struct {
		Current int `json:"current"`
		Total   int `json:"total"`
	} `json:"progressDetail"`
}

func PullImage(ctx context.Context, cli *client.Client, refStr string, auhthFunc types.RequestPrivilegeFunc, progressFunc func(progress float64)) error {
	log := log.FromContext(ctx)
	events, err := cli.ImagePull(ctx, refStr, types.ImagePullOptions{})
	if err != nil && strings.Contains(err.Error(), "no basic auth credentials") {
		var auth string
		auth, err = auhthFunc()
		if err != nil {
			log.Error(err, "failed to get auth credentials")
			return err
		}
		events, err = cli.ImagePull(ctx, refStr, types.ImagePullOptions{RegistryAuth: auth})
	}
	if err != nil {
		log.V(1).Error(err, "failed to pull image")
	}
	if events == nil {
		return err
	}
	defer events.Close()
	d := json.NewDecoder(events)
	layers := make(map[int]int)
	var percent float64
	var event *PullEvent
	for {
		if err := d.Decode(&event); err != nil {
			if err == io.EOF {
				break
			}
			log.Error(err, "error decoding image pull event")
			return err
		}
		if event.Error != "" {
			return fmt.Errorf("error pulling image: %s", event.Error)
		}
		log.V(1).Info(event.Progress)
		if event.ProgressDetail.Total != 0 {
			layers[event.ProgressDetail.Total] = event.ProgressDetail.Current
		}
		if len(layers) > 0 {
			var total, current float64
			for t, c := range layers {
				total += float64(t)
				current += float64(c)
			}
			percent = current / total * 100.0
		}
		progressFunc(percent)
	}

	// Latest event for new image
	// EVENT: {Status:Status: Downloaded newer image for busybox:latest Error: Progress:[==================================================>]  699.2kB/699.2kB ProgressDetail:{Current:699243 Total:699243}}
	// Latest event for up-to-date image
	// EVENT: {Status:Status: Image is up to date for busybox:latest Error: Progress: ProgressDetail:{Current:0 Total:0}}
	if event != nil {
		if strings.Contains(event.Status, fmt.Sprintf("Downloaded newer image for %s", refStr)) {
			// new
			log.Info("image pulled", "image", refStr)
		}

		if strings.Contains(event.Status, fmt.Sprintf("Image is up to date for %s", refStr)) {
			// up-to-date
			log.Info("image is up-to-date", "image", refStr)
		}
	}

	return err
}
