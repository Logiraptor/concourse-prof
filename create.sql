drop table if exists log_events;

create table log_events (
origin_id text,
origin_source text,
timestamp timestamp,
payload text
);

