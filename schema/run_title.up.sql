-- title is the last title the run's terminal set for itself, with an OSC 0
-- or 2 sequence such as printf '\033]0;building images\007'; empty when it
-- set none or cleared it.
ALTER TABLE run ADD COLUMN title TEXT NOT NULL DEFAULT '';
