package client

import (
	"encoding/json"
	"net/url"

	"golang.org/x/net/context"

	"github.com/docker/engine-api/types"
)

// ContainerWait pauses execution until a container exits.
// It returns the API status code as response of its readiness.
func (cli *Client) ContainerWait(ctx context.Context, containerID, custom string) (int, error) {
	query := url.Values{}
	if len(custom) > 0 {
		query.Set("custom", custom)
	}
	resp, err := cli.post(ctx, "/containers/"+containerID+"/wait", query, nil, nil)
	if err != nil {
		return -1, err
	}
	defer ensureReaderClosed(resp)

	var res types.ContainerWaitResponse
	if err := json.NewDecoder(resp.body).Decode(&res); err != nil {
		return -1, err
	}

	return res.StatusCode, nil
}
