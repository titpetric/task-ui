package runner

import (
	"context"
	"log/slog"
	"time"

	"github.com/titpetric/task-ui/model"
)

// enqueue puts a run whose job is running in the job's queue, queued,
// behind the runs that wait already, at most MaxQueued; a job that ended
// meanwhile starts the run at once.
func (m *Manager) enqueue(ctx context.Context, req Request, run *model.Run) (*Session, error) {
	key := run.Key()
	m.mu.Lock()
	full := len(m.queued[key]) >= MaxQueued
	m.mu.Unlock()
	if full {
		return nil, ErrQueueFull
	}
	run.State = model.StateQueued
	if err := m.opts.Runs.Create(ctx, run); err != nil {
		return nil, err
	}
	session := m.newSession(ctx, req, run)
	session.status = Status{State: model.StateQueued}

	m.mu.Lock()
	m.sessions[run.ID] = session
	if _, busy := m.byKey[key]; !busy {
		m.byKey[key] = session
		m.mu.Unlock()
		m.launch(session)
		return session, nil
	}
	if len(m.queued[key]) >= MaxQueued {
		delete(m.sessions, run.ID)
		m.mu.Unlock()
		_ = m.opts.Runs.Delete(ctx, run.ID)
		return nil, ErrQueueFull
	}
	m.queued[key] = append(m.queued[key], session)
	m.mu.Unlock()
	m.publish(session.Snapshot())
	return session, nil
}

// launch starts a session that waited: the row says it runs from now.
func (m *Manager) launch(session *Session) {
	now := time.Now().UTC().Truncate(time.Millisecond)
	session.mu.Lock()
	session.Run.State, session.Run.StartedAt = model.StateRunning, &now
	session.status = Status{State: model.StateRunning}
	run := *session.Run
	session.mu.Unlock()
	if err := m.opts.Runs.Update(context.Background(), &run); err != nil {
		slog.Warn("queued run not marked running", "run", run.ID, "err", err)
	}
	m.publish(run)
	go m.serve(session)
}

// handOff frees the lock of a key a session ends, or passes the lock to the
// first run that waits for it, which starts.
func (m *Manager) handOff(key string, session *Session) {
	m.mu.Lock()
	if m.byKey[key] != session {
		m.mu.Unlock()
		return
	}
	queue := m.queued[key]
	if len(queue) == 0 || m.closing {
		delete(m.byKey, key)
		m.mu.Unlock()
		return
	}
	next := queue[0]
	if len(queue) == 1 {
		delete(m.queued, key)
	} else {
		m.queued[key] = queue[1:]
	}
	m.byKey[key] = next
	m.mu.Unlock()
	m.launch(next)
}

// dequeue takes a queued session out of its queue and ends it in the
// status, and reports whether it waited: a run that started is none.
func (m *Manager) dequeue(session *Session, status Status) bool {
	key := session.Key()
	m.mu.Lock()
	queue := m.queued[key]
	at := -1
	for i, waiting := range queue {
		if waiting == session {
			at = i
		}
	}
	if at < 0 {
		m.mu.Unlock()
		return false
	}
	queue = append(queue[:at:at], queue[at+1:]...)
	if len(queue) == 0 {
		delete(m.queued, key)
	} else {
		m.queued[key] = queue
	}
	m.mu.Unlock()
	session.mu.Lock()
	session.Run.Message = status.Message
	session.mu.Unlock()
	m.settle(context.Background(), session, status)
	go m.forget(session)
	return true
}
