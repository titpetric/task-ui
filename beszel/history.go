package beszel

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// HistoryTTL is how long a month of records stays good: the coarsest
// record is eight hours long, so ten minutes miss little.
const HistoryTTL = 10 * time.Minute

// Month is how far back the history reaches: the hub keeps its eight
// hour records thirty days.
const Month = 30 * 24 * time.Hour

// historyTypes are the record types the history reads, coarsest first:
// the eight hour records cover the month in 90 points, and a hub younger
// than a record of them has the two hour ones.
var historyTypes = []string{"480m", "120m"}

// perPage is how many records a page of a list holds.
const perPage = 500

// record is a row of system_stats or container_stats: the host, when the
// hub wrote it, and its measures, decoded by the caller.
type record struct {
	System  string          `json:"system"`
	Created string          `json:"created"`
	Stats   json.RawMessage `json:"stats"`
}

// History returns every host's records of the month, oldest first, by
// the host's ID, at most HistoryTTL old.
func (c *Client) History(ctx context.Context) (map[string][]Sample, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.history != nil && time.Since(c.historyAt) < HistoryTTL {
		return c.history, nil
	}
	records, err := c.month(ctx, "system_stats", "")
	if err != nil {
		return nil, err
	}
	history := map[string][]Sample{}
	for _, record := range records {
		var stats Stats
		at, ok := createdAt(record.Created)
		if !ok || json.Unmarshal(record.Stats, &stats) != nil {
			continue
		}
		history[record.System] = append(history[record.System], Sample{At: at, Stats: stats})
	}
	c.history, c.historyAt = history, time.Now()
	return history, nil
}

// ContainerHistory returns a host's records of its containers of the
// month, oldest first, at most HistoryTTL old.
func (c *Client) ContainerHistory(ctx context.Context, system string) ([]ContainerSample, error) {
	if !validID(system) {
		return nil, nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if kept, ok := c.containerHistory[system]; ok && time.Since(kept.at) < HistoryTTL {
		return kept.samples, nil
	}
	records, err := c.month(ctx, "container_stats", system)
	if err != nil {
		return nil, err
	}
	samples := []ContainerSample{}
	for _, record := range records {
		var stats []containerStat
		at, ok := createdAt(record.Created)
		if !ok || json.Unmarshal(record.Stats, &stats) != nil {
			continue
		}
		sample := ContainerSample{At: at}
		for _, stat := range stats {
			sample.Containers = append(sample.Containers, stat.container())
		}
		samples = append(samples, sample)
	}
	if c.containerHistory == nil {
		c.containerHistory = map[string]containerHistory{}
	}
	c.containerHistory[system] = containerHistory{samples: samples, at: time.Now()}
	return samples, nil
}

// containerHistory is a host's container records as the client keeps
// them, and when it read them.
type containerHistory struct {
	samples []ContainerSample
	at      time.Time
}

// month lists a collection's records of the last Month, oldest first, of
// one host or of all: of the coarsest type the hub has any of.
func (c *Client) month(ctx context.Context, collection, system string) ([]record, error) {
	since := time.Now().UTC().Add(-Month).Format("2006-01-02 15:04:05")
	for _, kind := range historyTypes {
		filter := "type='" + kind + "' && created>='" + since + "'"
		if system != "" {
			filter += " && system='" + system + "'"
		}
		query := url.Values{"filter": {filter}, "sort": {"created"}, "fields": {"system,created,stats"}}
		records, err := c.list(ctx, collection, query)
		if err != nil || len(records) > 0 {
			return records, err
		}
	}
	return nil, nil
}

// list reads every page of a collection's records for a query.
func (c *Client) list(ctx context.Context, collection string, query url.Values) ([]record, error) {
	query.Set("perPage", strconv.Itoa(perPage))
	query.Set("skipTotal", "1")
	var all []record
	for page := 1; ; page++ {
		query.Set("page", strconv.Itoa(page))
		var list struct {
			Items []record `json:"items"`
		}
		if _, err := c.do(ctx, "/api/collections/"+collection+"/records?"+query.Encode(), &list); err != nil {
			return nil, err
		}
		all = append(all, list.Items...)
		if len(list.Items) < perPage {
			return all, nil
		}
	}
}

// createdAt reads a record's created field, in the layout PocketBase writes.
func createdAt(created string) (time.Time, bool) {
	for _, layout := range []string{"2006-01-02 15:04:05.000Z", "2006-01-02 15:04:05Z", time.RFC3339Nano} {
		if at, err := time.Parse(layout, created); err == nil {
			return at, true
		}
	}
	return time.Time{}, false
}

// validID reports whether an ID is one PocketBase makes, safe in a
// filter.
func validID(id string) bool {
	return id != "" && strings.Trim(id, "abcdefghijklmnopqrstuvwxyz0123456789") == ""
}
