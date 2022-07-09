--
-- Initial SQL scripts for schemas which the application uses
--

CREATE TABLE membership_warning (
    warning_id integer NOT NULL,
    user_id bigint NOT NULL,
    username name NOT NULL,
    first_warning_ts timestamp with time zone NOT NULL,
    grace_period varchar NOT NULL,
    is_relevant boolean NOT NULL
);

ALTER TABLE membership_warning ALTER COLUMN warning_id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME membership_warning_warning_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);

CREATE INDEX membership_warning_relevant_index ON membership_warning (is_relevant)
    WHERE is_relevant is true;

ALTER TABLE ONLY membership_warning
    ADD CONSTRAINT membership_warning_pkey PRIMARY KEY (warning_id);

CREATE TABLE phrase_type (
    type_id integer NOT NULL,
    name name NOT NULL
);

ALTER TABLE ONLY phrase_type
    ADD CONSTRAINT phrase_type_pkey PRIMARY KEY (type_id);

CREATE TABLE phrase (
    phrase_id integer NOT NULL,
    text text NOT NULL,
    is_user_templated boolean NOT NULL,
    weight integer NOT NULL,
    type integer NOT NULL,
    is_audio_accompaniment boolean NOT NULL,
    vk_audio_id character varying(64)
);

ALTER TABLE phrase ALTER COLUMN phrase_id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME phrase_phrase_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);

ALTER TABLE ONLY phrase
    ADD CONSTRAINT phrase_pkey PRIMARY KEY (phrase_id);

ALTER TABLE ONLY phrase
    ADD CONSTRAINT phrase FOREIGN KEY (type) REFERENCES phrase_type(type_id);

CREATE TABLE content_command_type (
     id integer NOT NULL,
     name name NOT NULL
);

ALTER TABLE ONLY content_command_type
    ADD CONSTRAINT content_command_type_pkey PRIMARY KEY (id);

CREATE TABLE content_command (
    id integer NOT NULL,
    name varchar NOT NULL,
    media_type integer NOT NULL,
    community_ids text NOT NULL
);

ALTER TABLE content_command ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME content_command_id_seq
        START WITH 1
        INCREMENT BY 1
        NO MINVALUE
        NO MAXVALUE
        CACHE 1
    );

ALTER TABLE ONLY content_command
    ADD CONSTRAINT content_command_pkey PRIMARY KEY (id);

ALTER TABLE ONLY content_command
    ADD CONSTRAINT content_command FOREIGN KEY (media_type) REFERENCES content_command_type(id);

CREATE UNIQUE INDEX uidx_name
    ON content_command (name);

--
-- Initial SQL scripts for needed data which the application uses to work by default
--

-- Used phrase types in the application
INSERT INTO phrase_type (type_id, name)  VALUES (1, 'welcome');
INSERT INTO phrase_type (type_id, name)  VALUES (2, 'goodbye');
INSERT INTO phrase_type (type_id, name)  VALUES (3, 'membership_warning');
INSERT INTO phrase_type (type_id, name)  VALUES (4, 'info');
INSERT INTO phrase_type (type_id, name)  VALUES (5, 'content_request');

-- Used content source types in the application
INSERT INTO content_command_type (id, name)  VALUES (1, 'audio');
INSERT INTO content_command_type (id, name)  VALUES (2, 'picture');
