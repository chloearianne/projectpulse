DROP DATABASE IF EXISTS ppdb;
CREATE DATABASE ppdb;

DROP OWNED BY ppmaster CASCADE;
REVOKE ALL ON ALL TABLES IN SCHEMA public FROM PUBLIC;

DROP USER IF EXISTS ppmaster;
CREATE USER ppmaster;

ALTER DATABASE ppdb OWNER TO ppmaster;
GRANT ALL PRIVILEGES ON DATABASE ppdb TO ppmaster;

-- Connect to newly created database
\c ppdb ppmaster;

CREATE TABLE account (
    id SERIAL PRIMARY KEY,
    email varchar NOT NULL,
    password varchar NOT NULL,
    first_name varchar NOT NULL,
    last_name varchar NOT NULL,
    CONSTRAINT uniq_voter_email UNIQUE(email)
);

CREATE TABLE event_type (
    id    SERIAL PRIMARY KEY,
    name  varchar,
    CONSTRAINT uniq_type UNIQUE(name)
);

INSERT INTO event_type (name) VALUES
    ('in person'),
    ('online'),
    ('donation');

CREATE TABLE event_topic (
    id    SERIAL PRIMARY KEY,
    name  varchar,
    CONSTRAINT uniq_topic UNIQUE(name)
);

INSERT INTO event_topic (name) VALUES
    ('police violence'),
    ('environment'),
    ('gender equality'),
    ('racial justice'),
    ('lgbtq rights'),
    ('indigenous rights'),
    ('animal rights'),
    ('other');

CREATE TABLE event (
    id               SERIAL PRIMARY KEY,
    creator_id       integer REFERENCES account ON DELETE CASCADE,
    title            varchar,
    start_timestamp  timestamp,
    end_timestamp    timestamp,
    description      text,
    event_type       integer REFERENCES event_type ON DELETE CASCADE,
    event_topic      integer REFERENCES event_topic ON DELETE CASCADE,
    location         varchar,
    stars            integer
);

CREATE TABLE account_event_topics (
    account_id       integer REFERENCES account ON DELETE CASCADE,
    topic_id      integer REFERENCES event_topic ON DELETE CASCADE,
    PRIMARY KEY(account_id, topic_id)
);

CREATE TABLE account_event_type (
    account_id       integer REFERENCES account ON DELETE CASCADE,
    type_id       integer REFERENCES event_type ON DELETE CASCADE,
    PRIMARY KEY(account_id, type_id)
);

CREATE TABLE account_star (
    account_id   integer REFERENCES account ON DELETE CASCADE,
    event_id  integer REFERENCES event ON DELETE CASCADE,
    PRIMARY KEY(account_id, event_id)
);
