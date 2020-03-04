CREATE TABLE record (
  id serial PRIMARY KEY,
  uid varchar(255) NOT NULL,
  ts timestamp,
  data json
);

CREATE INDEX record_ts_uid ON record USING btree(ts,uid);

CREATE TABLE installation (
  id serial PRIMARY KEY,
  uid varchar(255) UNIQUE NOT NULL,
  first_seen timestamp,
  last_seen timestamp,
  last_ip varchar(255),
  last_record int REFERENCES record(id),
  note text
);

CREATE INDEX installation_last_seen ON installation USING btree(last_seen);

CREATE TABLE byday (
  id serial PRIMARY KEY,
  uid varchar(255) NOT NULL,
  day date NOT NULL,
  record_id int REFERENCES record(id)
);

CREATE UNIQUE INDEX byday_day_uid ON byday USING btree(day,uid);

CREATE TABLE account (
  id serial PRIMARY KEY,
  name varchar(255) NOT NULL UNIQUE,
  hash varchar(255)
);
