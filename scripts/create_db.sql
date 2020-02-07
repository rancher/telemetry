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

CREATE TABLE license_installation_record (
  id serial PRIMARY KEY,
  uid varchar(255) NOT NULL,
  license_key varchar(255) NOT NULL,
  telemetry_uid varchar(255) NOT NULL,
  ts timestamp,
  data json
);

CREATE INDEX license_installation_record_uid ON license_installation_record USING btree(uid);

CREATE INDEX license_installation_record_ts_uid ON license_installation_record USING btree(ts,uid);

CREATE TABLE license_record (
  id serial PRIMARY KEY,
  key varchar(255)  NOT NULL,
  ts timestamp,
  data json
);

CREATE INDEX license_record_key ON license_record USING btree(key);

CREATE INDEX license_record_ts_key ON license_record USING btree(ts,key);

CREATE TABLE license (
  id serial PRIMARY KEY,
  key varchar(255) UNIQUE NOT NULL,
  first_seen timestamp,
  last_ip varchar(255),
  last_seen timestamp,
  last_record int REFERENCES license_record(id),
  licensed_installations int NOT NULL,
  licensed_nodes int NOT NULL,
  valid_from timestamp NOT NULL,
  valid_to timestamp NOT NULL,
  note text
);

CREATE INDEX license_uid ON license USING btree(uid);

CREATE INDEX license_last_seen ON license USING btree(last_seen);

CREATE TABLE license_installation (
  id serial PRIMARY KEY,
  uid varchar(255) UNIQUE NOT NULL,
  license_key varchar(255) REFERENCES license(key),
  telemetry_uid varchar(255) NOT NULL,
  first_seen timestamp,
  last_seen timestamp,
  last_ip varchar(255),
  last_record int REFERENCES license_installation_record(id),
  note text,
  running_nodes int NOT NULL
);

CREATE INDEX license_installation_uid ON license_installation USING btree(uid);

CREATE INDEX license_installation_last_seen ON license_installation USING btree(last_seen);

