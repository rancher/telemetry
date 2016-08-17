CREATE TABLE record (
  id serial PRIMARY KEY,
  uid varchar(255) NOT NULL,
  ts timestamp,
  data json
);

CREATE INDEX record_uid ON record USING btree(uid);

CREATE TABLE installation (
  id serial PRIMARY KEY,
  uid varchar(255) UNIQUE NOT NULL,
  first_seen timestamp,
  last_seen timestamp,
  last_ip varchar(255),
  last_record int REFERENCES record(id)
);
