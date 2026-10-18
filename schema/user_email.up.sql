-- user_email: An email address a user signs in with, one user may have several
CREATE TABLE IF NOT EXISTS user_email (
    id CHAR(26) PRIMARY KEY NOT NULL,
    user_id CHAR(26) NOT NULL,
    email TEXT NOT NULL,
    created_at DATETIME,
    updated_at DATETIME
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_email_email ON user_email(email);
CREATE INDEX IF NOT EXISTS idx_user_email_user_id ON user_email(user_id);
