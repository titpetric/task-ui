-- root_only lists a project's own jobs alone: the folders inside it are
-- not discovered. Off, the default, discovery goes through the checkout.
ALTER TABLE project ADD COLUMN root_only BOOLEAN NOT NULL DEFAULT 0;
