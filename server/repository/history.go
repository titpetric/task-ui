package repository

import (
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/apex/log"

	. "github.com/titpetric/task-ui/server/model"
)

func FillHistory(h *HistoryResponse, files []string) {
	split := func(in string) (string, string) {
		v := strings.Split(in, "-")
		return v[0], v[1]
	}

	for _, file := range files {
		base := path.Base(file)
		ext := path.Ext(file)
		root := base[:len(base)-len(ext)]

		id, timestamp := split(root)
		ts, _ := strconv.ParseInt(timestamp, 10, 64)
		t := time.Unix(ts, 0)

		record := HistoryRecord{
			ID:        root,
			Timestamp: ts,
			Datetime:  t.Format(time.RFC3339Nano),
			Seconds:   0,
			Lines:     0,
		}

		seconds, err := passthru("ttytime " + file)
		if err != nil {
			log.Warnf("Error getting duration for %s", file)
		}
		record.Seconds, err = strconv.Atoi(seconds)

		history, ok := h.History[id]
		if !ok {
			h.History[id] = []HistoryRecord{record}
		} else {
			history = append(history, record)
			h.History[id] = history
		}
	}
}

func LoadHistory(name string) (result []HistoryRecord) {
	files, err := filepath.Glob("history/" + name + "-*.ttyrec")
	if err != nil {
		return
	}

	response := NewHistoryResponse()

	FillHistory(response, files)

	return response.History[name]
}
