DROP TABLE IF EXISTS emails;

CREATE TABLE emails (
    email_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    "to" VARCHAR(500) NOT NULL,
    "from" VARCHAR(250) NOT NULL,
    subject VARCHAR(250) NOT NULL,
    body VARCHAR(250) NOT NULL,
    content_type VARCHAR(250) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);