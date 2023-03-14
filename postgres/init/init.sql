CREATE TABLE IF NOT EXISTS template_content (
    id      SERIAL PRIMARY KEY,
    content TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS template (
    id      SERIAL PRIMARY KEY,
    name    TEXT UNIQUE NOT NULL,
    content INT REFERENCES template_content ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS certificate (
    id          TEXT UNIQUE NOT NULL,
    template    INT REFERENCES template ON DELETE RESTRICT,
    timestamp   TIMESTAMP,
    student     TEXT,
    issue_date  TEXT,
    course      TEXT,
    mentors     TEXT
);

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE OR REPLACE FUNCTION generate_id() RETURNS TRIGGER AS $generate_id$
    DECLARE
        new_id  TEXT;
        counter     INT;
    BEGIN
        counter = 0;
        LOOP
            new_id = encode(gen_random_bytes(4), 'hex');
            IF (SELECT id FROM certificate WHERE id=new_id) IS NULL THEN
                NEW.id = new_id;
                RETURN NEW;
            END IF;
            counter = counter + 1;
            IF counter = 10 THEN
                RAISE EXCEPTION 'ID generation failed. Retry limit exceeded';
            END IF;
        END LOOP;
    END
$generate_id$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER set_id
    BEFORE INSERT ON certificate
    FOR EACH ROW EXECUTE PROCEDURE generate_id();

CREATE OR REPLACE FUNCTION update_timestamp() RETURNS TRIGGER AS $update_timestamp$
    BEGIN
	IF TG_TABLE_NAME = 'certificate' THEN
		NEW.timestamp = now();
		RETURN NEW;
	--Lefted for possible use.
	--ELSIF TG_TABLE_NAME = 'template' THEN
	--	UPDATE certificate SET timestamp = now()
	--	WHERE template = NEW.id;
	ELSIF TG_TABLE_NAME = 'template_content' THEN
		UPDATE certificate SET timestamp = now() FROM template
		WHERE template.content = NEW.id AND certificate.template = template.id ;
	END IF;
	RETURN NULL;
    END;
$update_timestamp$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER update_timestamp_certificate
    BEFORE INSERT OR UPDATE ON certificate
    FOR EACH ROW EXECUTE PROCEDURE update_timestamp();

--CREATE OR REPLACE TRIGGER update_timestamp_template
--    AFTER UPDATE ON template
--    FOR EACH ROW EXECUTE PROCEDURE update_timestamp();

CREATE OR REPLACE TRIGGER update_timestamp_template_content
    AFTER UPDATE ON template_content
    FOR EACH ROW EXECUTE PROCEDURE update_timestamp();
