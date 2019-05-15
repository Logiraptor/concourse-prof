drop table if exists log_events;
drop table if exists status_events;

create table log_events (
  pipeline text,
  job text,
  build text,
  origin_id text,
  origin_source text,
  timestamp timestamp,
  payload text
);

create table status_events (
  pipeline text,
  job text,
  build text,
  status text,
  timestamp timestamp
);
