-- +goose Up
CREATE TABLE current (
   time        TIMESTAMP (0) WITHOUT TIME ZONE NOT NULL,
   phase       INTEGER NOT NULL DEFAULT(0),
   current     INTEGER NOT NULL,
   UNIQUE (time, phase)
);

CREATE TABLE power (
   time        TIMESTAMP (0) WITHOUT TIME ZONE UNIQUE NOT NULL,
   power       INTEGER NOT NULL
);

CREATE TABLE energy (
   time        TIMESTAMP (0) WITHOUT TIME ZONE NOT NULL,
   tariff      TEXT NOT NULL,
   reading     INTEGER NOT NULL,
   UNIQUE (time, tariff)
);

SELECT create_hypertable('current','time');
SELECT create_hypertable('power','time');
SELECT create_hypertable('energy','time');

-- +goose Down
DROP TABLE current;
DROP TABLE power;
DROP TABLE energy;
