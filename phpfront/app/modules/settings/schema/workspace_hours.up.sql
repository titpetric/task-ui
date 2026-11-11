-- hours says when the runs of the workspace's folders may start, as the
-- hours package reads them, one rule per line; '', the default, allows
-- any time. A run outside them needs Force and a reason.
ALTER TABLE workspace ADD COLUMN hours TEXT NOT NULL DEFAULT '';
