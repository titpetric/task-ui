package beszel

import (
	"context"
	"encoding/json"
	"net/url"
	"time"
)

// Containers returns the containers each host runs, by the host's ID, at most TTL old.
// They come from the hub's containers collection, which holds what its agents reported in the last ten minutes.
// On a hub without that collection, they are the newest container_stats record of each host.
func (c *Client) Containers(ctx context.Context) (map[string][]Container, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.containers != nil && time.Since(c.containersAt) < c.TTL {
		return c.containers, nil
	}
	containers, err := c.containerList(ctx)
	if err != nil {
		containers, err = c.containerStats(ctx)
	}
	if err != nil {
		return nil, err
	}
	c.containers, c.containersAt = containers, time.Now()
	return containers, nil
}

// containerRecord is a row of the containers collection.
type containerRecord struct {
	System string `json:"system"`
	Container
}

// containerList reads the containers collection.
func (c *Client) containerList(ctx context.Context) (map[string][]Container, error) {
	query := url.Values{"sort": {"name"}, "fields": {"system,name,image,status,cpu,memory,net"}, "perPage": {"1000"}, "skipTotal": {"1"}}
	var list struct {
		Items []containerRecord `json:"items"`
	}
	if _, err := c.do(ctx, "/api/collections/containers/records?"+query.Encode(), &list); err != nil {
		return nil, err
	}
	containers := map[string][]Container{}
	for _, record := range list.Items {
		containers[record.System] = append(containers[record.System], record.Container)
	}
	return containers, nil
}

// containerStats reads the newest minute of container_stats of each host
// of the last ten minutes.
func (c *Client) containerStats(ctx context.Context) (map[string][]Container, error) {
	since := time.Now().UTC().Add(-10 * time.Minute).Format("2006-01-02 15:04:05")
	query := url.Values{"filter": {"type='1m' && created>='" + since + "'"}, "sort": {"-created"}, "fields": {"system,created,stats"}}
	records, err := c.list(ctx, "container_stats", query)
	if err != nil {
		return nil, err
	}
	containers := map[string][]Container{}
	for _, record := range records {
		if _, ok := containers[record.System]; ok {
			continue
		}
		var stats []containerStat
		if json.Unmarshal(record.Stats, &stats) != nil {
			continue
		}
		containers[record.System] = []Container{}
		for _, stat := range stats {
			containers[record.System] = append(containers[record.System], stat.container())
		}
	}
	return containers, nil
}
