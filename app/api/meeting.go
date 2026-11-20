package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"golang.org/x/net/websocket"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/feature/meeting"
	"github.com/titpetric/task-ui/internal/access"
	"github.com/titpetric/task-ui/internal/markdown"
)

// meetingHeartbeat is how often a pipe records its user as still in the room.
const meetingHeartbeat = 10 * time.Second

// meetingOff answers a request on an instance without the room service.
func (h *Handlers) meetingOff(w http.ResponseWriter) bool {
	if h.opts.Meetings == nil {
		fail(w, http.StatusNotFound, "meetings aren't available on this instance")
		return true
	}
	return false
}

// meetingView reads the room a request names, answering an error when it is not there.
func (h *Handlers) meetingView(w http.ResponseWriter, r *http.Request) (*meeting.View, bool) {
	view, err := h.opts.Meetings.Get(r.Context(), param(r, "id"))
	if err != nil {
		failErr(w, err)
		return nil, false
	}
	return view, true
}

// listMeetings returns the rooms, open ones first, at most limit.
func (h *Handlers) listMeetings(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "meeting/view") || h.meetingOff(w) {
		return
	}
	rooms, err := h.opts.Meetings.List(r.Context(), queryInt(r, "limit", 20))
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"meetings": rooms})
}

// createMeeting opens a room as the user, titled by the body.
func (h *Handlers) createMeeting(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "meeting/edit") || h.meetingOff(w) {
		return
	}
	var req struct {
		Title string `json:"title"`
	}
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	room, err := h.opts.Meetings.Create(r.Context(), access.UserFrom(r.Context()), req.Title)
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"meeting": room})
}

// getMeeting returns one room with its participants and its log.
func (h *Handlers) getMeeting(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "meeting/view") || h.meetingOff(w) {
		return
	}
	view, ok := h.meetingView(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, viewJSON(view))
}

// deleteMeeting removes a room with its log and its audio.
func (h *Handlers) deleteMeeting(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "meeting/delete") || h.meetingOff(w) {
		return
	}
	view, ok := h.meetingView(w, r)
	if !ok {
		return
	}
	if user := access.UserFrom(r.Context()); view.Meeting.UserID != user.ID && !user.CanAdmin() {
		fail(w, http.StatusForbidden, "only the creator or an admin deletes the meeting")
		return
	}
	if err := h.opts.Meetings.Delete(r.Context(), param(r, "id")); err != nil {
		failErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// joinMeeting puts the user in a room, logging join or rejoin.
func (h *Handlers) joinMeeting(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "meeting/edit") || h.meetingOff(w) {
		return
	}
	room, participant, line, err := h.opts.Meetings.Join(r.Context(), access.UserFrom(r.Context()), param(r, "id"))
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"meeting": room, "participant": participant, "line": line})
}

// leaveMeeting takes the user out of a room.
func (h *Handlers) leaveMeeting(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "meeting/edit") || h.meetingOff(w) {
		return
	}
	line, err := h.opts.Meetings.Leave(r.Context(), access.UserFrom(r.Context()), param(r, "id"))
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"line": line})
}

// heartbeatMeeting marks the user as still in a room.
func (h *Handlers) heartbeatMeeting(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "meeting/edit") || h.meetingOff(w) {
		return
	}
	if err := h.opts.Meetings.Heartbeat(r.Context(), access.UserFrom(r.Context()), param(r, "id")); err != nil {
		failErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// finishMeeting ends a room as its creator or an admin and starts its summary.
func (h *Handlers) finishMeeting(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "meeting/edit") || h.meetingOff(w) {
		return
	}
	room, err := h.opts.Meetings.Finish(r.Context(), access.UserFrom(r.Context()), param(r, "id"))
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"meeting": room})
}

// setMeetingAgent names the room's agent, or clears it, as its creator or an admin.
func (h *Handlers) setMeetingAgent(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "meeting/edit") || h.meetingOff(w) {
		return
	}
	var req struct {
		Agent string `json:"agent"`
	}
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	room, err := h.opts.Meetings.SetAgent(r.Context(), access.UserFrom(r.Context()), param(r, "id"), req.Agent)
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"meeting": room})
}

// meetingSummary returns the summary an ended room has.
func (h *Handlers) meetingSummary(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "meeting/view") || h.meetingOff(w) {
		return
	}
	view, ok := h.meetingView(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, meetingSummaryJSON(view.Meeting))
}

// summarizeMeeting asks Claude again for the summary of an ended room.
func (h *Handlers) summarizeMeeting(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "meeting/edit") || h.meetingOff(w) {
		return
	}
	room, err := h.opts.Meetings.RetrySummary(r.Context(), access.UserFrom(r.Context()), param(r, "id"))
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, meetingSummaryJSON(room))
}

// meetingChannel serves one participant's speech as a WAV, with Range support.
func (h *Handlers) meetingChannel(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "meeting/view") || h.meetingOff(w) {
		return
	}
	reader, _, err := h.opts.Meetings.Channels.WAV(param(r, "id"), param(r, "user"))
	if err != nil {
		failErr(w, err)
		return
	}
	defer reader.(io.Closer).Close()
	w.Header().Set("Content-Type", "audio/wav")
	w.Header().Set("Cache-Control", "no-store")
	http.ServeContent(w, r, param(r, "user")+".wav", time.Time{}, reader)
}

// meetingPipe serves the websocket a participant speaks and listens on.
func (h *Handlers) meetingPipe() http.Handler {
	return websocket.Server{
		Handshake: func(config *websocket.Config, r *http.Request) error {
			origin, err := url.Parse(r.Header.Get("Origin"))
			if err != nil || !access.SameHost(r, origin.Host) {
				return errors.New("cross origin websocket refused")
			}
			config.Origin = origin
			return nil
		},
		Handler: func(ws *websocket.Conn) {
			defer ws.Close()
			h.serveMeetingPipe(ws)
		},
	}
}

// serveMeetingPipe relays a room's audio and control messages between its participants.
func (h *Handlers) serveMeetingPipe(ws *websocket.Conn) {
	r := ws.Request()
	user := access.UserFrom(r.Context())
	if h.opts.Meetings == nil {
		sendMeeting(ws, meeting.Message{Type: "error", Message: "meetings aren't available on this instance"})
		return
	}
	if !h.can(r, "meeting/edit") {
		sendMeeting(ws, meeting.Message{Type: "error", Message: "you need the meeting/edit permission"})
		return
	}
	id := param(r, "id")
	if _, _, _, err := h.opts.Meetings.Join(r.Context(), user, id); err != nil {
		sendMeeting(ws, meeting.Message{Type: "error", Message: err.Error()})
		return
	}
	view, err := h.opts.Meetings.Get(r.Context(), id)
	if err != nil {
		sendMeeting(ws, meeting.Message{Type: "error", Message: err.Error()})
		return
	}
	pipe := &meeting.Pipe{MeetingID: id, UserID: user.ID, UserName: user.Name, Frames: make(chan meeting.Frame, meeting.Buffer)}
	h.opts.Meetings.Hub.Join(pipe)
	// A closed socket leaves presence to the sweeper: a reload comes back
	// inside StaleAfter, and logs no part and no rejoin.
	defer h.opts.Meetings.Hub.Leave(pipe)
	sendMeeting(ws, meeting.Message{Type: "hello", Meeting: view.Meeting, Participants: view.Participants, Lines: view.Lines})

	gone := make(chan struct{})
	var once sync.Once
	go func() {
		defer once.Do(func() { close(gone) })
		h.readMeeting(ws, pipe, view.Meeting, user)
	}()
	for {
		select {
		case <-gone:
			return
		case frame, ok := <-pipe.Frames:
			if !ok {
				return
			}
			if len(frame.Text) > 0 && websocket.Message.Send(ws, string(frame.Text)) != nil {
				return
			}
			if len(frame.Data) > 0 && websocket.Message.Send(ws, frame.Data) != nil {
				return
			}
		}
	}
}

// readMeeting applies what a participant sends until the socket closes.
func (h *Handlers) readMeeting(ws *websocket.Conn, pipe *meeting.Pipe, room *model.Meeting, user *model.User) {
	ctx := ws.Request().Context()
	hub := h.opts.Meetings.Hub
	var frame wireFrame
	var seq int64
	seen := time.Now()
	for {
		if err := frameCodec.Receive(ws, &frame); err != nil {
			return
		}
		if time.Since(seen) >= meetingHeartbeat {
			_ = h.opts.Meetings.Heartbeat(ctx, user, room.ID)
			seen = time.Now()
		}
		switch frame.Kind {
		case websocket.BinaryFrame:
			if len(frame.Data) == 0 {
				continue
			}
			offset, err := h.opts.Meetings.Audio(ctx, room, user, frame.Data)
			if err != nil {
				return
			}
			seq++
			hub.Broadcast(pipe.MeetingID, pipe, meeting.AudioFrame(user.ID, user.Name, seq, offset/meeting.BytesPerMS, frame.Data))
		case websocket.TextFrame:
			var control pipeControl
			if json.Unmarshal(frame.Data, &control) != nil {
				continue
			}
			h.applyControl(ctx, pipe, room, user, control)
		}
	}
}

// applyControl logs a final transcript line or relays an interim one or a mute.
func (h *Handlers) applyControl(ctx context.Context, pipe *meeting.Pipe, room *model.Meeting, user *model.User, control pipeControl) {
	hub := h.opts.Meetings.Hub
	switch control.Type {
	case "line":
		line, err := h.opts.Meetings.Line(ctx, room, user, control.Text, control.DurationMS)
		if err == nil && line != nil {
			hub.Broadcast(pipe.MeetingID, nil, meeting.TextFrame(meeting.Message{Type: "line", Line: line}))
		}
	case "partial":
		hub.Broadcast(pipe.MeetingID, pipe, meeting.TextFrame(meeting.Message{Type: "partial", UserID: user.ID, UserName: user.Name, Text: control.Text}))
	case "mute":
		hub.Broadcast(pipe.MeetingID, pipe, meeting.TextFrame(meeting.Message{Type: "mute", UserID: user.ID, UserName: user.Name, Muted: control.On}))
	}
}

// pipeControl is what a window sends as text on the meeting socket.
type pipeControl struct {
	Type       string `json:"type"`
	Text       string `json:"text"`
	DurationMS int64  `json:"duration_ms"`
	On         bool   `json:"on"`
}

// wireFrame is one frame off a socket, its text or binary kind kept.
type wireFrame struct {
	Kind byte
	Data []byte
}

// frameCodec receives a frame and keeps the payload type Message.Receive drops.
var frameCodec = websocket.Codec{Unmarshal: unmarshalFrame}

// unmarshalFrame keeps the payload type Message.Receive drops.
func unmarshalFrame(data []byte, payloadType byte, v any) error {
	frame := v.(*wireFrame)
	frame.Kind, frame.Data = payloadType, data
	return nil
}

// sendMeeting writes one control message.
func sendMeeting(ws *websocket.Conn, msg meeting.Message) {
	_ = websocket.Message.Send(ws, string(meeting.TextFrame(msg).Text))
}

// viewJSON is a room with its participants and its log, as the API returns them.
func viewJSON(view *meeting.View) map[string]any {
	return map[string]any{"meeting": view.Meeting, "participants": view.Participants, "lines": view.Lines}
}

// meetingSummaryJSON is a room's summary as the API returns it: markdown, and the same as HTML.
func meetingSummaryJSON(room *model.Meeting) map[string]any {
	return map[string]any{
		"meeting_id": room.ID,
		"summary":    room.Summary,
		"html":       markdown.HTML([]byte(room.Summary), "", ""),
		"state":      room.SummaryState,
		"message":    room.SummaryMessage,
		"user_name":  room.SummaryUserName,
	}
}
