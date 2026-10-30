-- refinement_comment: what kind of comment it is, and which Apply all folded it
--
-- kind is '' for the two a person writes, the design and a correction,
-- which the correction row tells apart, and 'apply' for the record an
-- Apply all leaves: the prompt it sent. An apply comment is neither a
-- design nor a rule, so the prompt is never composed out of one.
ALTER TABLE refinement_comment ADD COLUMN kind TEXT NOT NULL DEFAULT '';

-- collapsed_into is the apply comment this one was folded into, and ''
-- while it stands on its own. A folded correction is still a rule: the
-- fold is what the page draws, and deleting is what retires a rule.
ALTER TABLE refinement_comment ADD COLUMN collapsed_into CHAR(26) NOT NULL DEFAULT '';
