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
	key := run.LockKey()
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
	session.slot = TakesSlot(req.Job)
	if _, busy := m.byKey[key]; !busy {
		m.byKey[key] = session
		m.mu.Unlock()
		m.admit(session)
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

// handOff frees the lock of a key a session ends, nil for a start that
// gave the lock up, or passes the lock to the first run that waits for
// it, which starts.
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
	m.admit(next)
}

// dequeue takes a queued session out of its queue, or out of the line
// for a slot, which frees its job for the run that waits next, and ends
// it in the status; it reports whether it waited: a run that started
// is none.
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
		if !m.slots.Drop(session.ID()) {
			return false
		}
		m.handOff(key, session)
		m.end(session, status)
		return true
	}
	queue = append(queue[:at:at], queue[at+1:]...)
	if len(queue) == 0 {
		delete(m.queued, key)
	} else {
		m.queued[key] = queue
	}
	m.mu.Unlock()
	m.end(session, status)
	return true
}

// end settles a session that never ran in the status.
func (m *Manager) end(session *Session, status Status) {
	session.mu.Lock()
	session.Run.Message = status.Message
	session.mu.Unlock()
	m.settle(context.Background(), session, status)
	go m.forget(session)
}

// admit starts a session that holds its job's lock: at once when it
// does no work that takes a slot or one is free, else it waits, queued,
// for a slot. A session that waited for its job starts running here.
func (m *Manager) admit(session *Session) {
	m.mu.Lock()
	going := !session.slot || m.slots.Take(session.ID())
	m.mu.Unlock()
	if !going {
		session.mu.Lock()
		marked := session.status.State == model.StateQueued
		session.Run.State, session.status = model.StateQueued, Status{State: model.StateQueued}
		run := *session.Run
		session.mu.Unlock()
		if !marked {
			if err := m.opts.Runs.Update(context.Background(), &run); err != nil {
				slog.Warn("run not marked queued", "run", run.ID, "err", err)
			}
		}
		m.publish(run)
		return
	}
	if session.Status().State == model.StateQueued {
		m.launch(session)
		return
	}
	m.publish(session.Snapshot())
	go m.serve(session)
}

// freeSlot releases the slot of a session that ended, and starts the
// sessions that waited first for one.
func (m *Manager) freeSlot(session *Session) {
	if !session.slot {
		return
	}
	m.mu.Lock()
	var next []*Session
	for _, id := range m.slots.Free(session.ID()) {
		if waiting, ok := m.sessions[id]; ok && !m.closing {
			next = append(next, waiting)
		}
	}
	m.mu.Unlock()
	for _, waiting := range next {
		m.launch(waiting)
	}
}
