CREATE TABLE membership_warning (
    warning_id integer NOT NULL,
    user_id bigint NOT NULL,
    username name NOT NULL,
    first_warning_ts timestamp with time zone NOT NULL,
    grace_period_ns bigint NOT NULL,
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

CREATE TABLE source_type (
     source_type_id integer NOT NULL,
     name name NOT NULL
);

ALTER TABLE ONLY source_type
    ADD CONSTRAINT source_type_pkey PRIMARY KEY (source_type_id);

CREATE TABLE content_source (
    source_id integer NOT NULL,
    vk_community_id character varying(64) NOT NULL,
    type integer NOT NULL
);

ALTER TABLE ONLY content_source
    ADD CONSTRAINT content_source_pkey PRIMARY KEY (source_id);

ALTER TABLE ONLY content_source
    ADD CONSTRAINT content_source FOREIGN KEY (type) REFERENCES source_type(source_type_id);