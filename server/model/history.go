package model

type HistoryResponse struct {
	Files   []string
	History map[string][]HistoryRecord
}

func NewHistoryResponse() *HistoryResponse {
	return &HistoryResponse{
		History: make(map[string][]HistoryRecord),
	}
}

type HistoryRecord struct {
	ID        string
	Timestamp int64
	Datetime  string
	Seconds   int
	Lines     int
}
