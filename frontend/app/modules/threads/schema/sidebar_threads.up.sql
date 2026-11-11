-- The Threads timeline takes the slot Runs held and the one Notes held: a
-- thread is a note, a prompt or a question put to the orb, and they are one
-- list. The runs pages answer where they did, from a thread's messages and
-- from a folder.
UPDATE sidebar SET href = '/threads', title = 'Threads', icon = 'note', updated_at = CURRENT_TIMESTAMP WHERE href = '/runs';

DELETE FROM sidebar WHERE href = '/notes';
