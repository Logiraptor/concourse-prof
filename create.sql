drop table if exists start_get_events;
drop table if exists finish_get_events;
drop table if exists initialize_get_events;
drop table if exists finish_put_events;
drop table if exists start_put_events;
drop table if exists initialize_put_events;
drop table if exists finish_task_events;
drop table if exists start_task_events;
drop table if exists initialize_task_events;
drop table if exists log_events;
drop table if exists status_events;

-- create table log_events (
--   pipeline text,
--   job text,
--   build text,
--   origin_id text,
--   origin_source text,
--   timestamp timestamp,
--   payload text
-- );

-- create table status_events (
--   pipeline text,
--   job text,
--   build text,
--   status text,
--   timestamp timestamp
-- );

-- create table initialize_task_events (
-- pipeline text,
-- job text,
-- build text,
-- timestamp timestamp,
-- origin_id text,
-- platform text,
-- image text,
-- run text,
-- dir text
-- );
-- create table start_task_events (
-- pipeline text,
-- job text,
-- build text,
-- timestamp timestamp,
-- origin_id text,
-- platform text,
-- image text,
-- run text,
-- dir text
-- );
-- create table finish_task_events (
-- pipeline text,
-- job text,
-- build text,
-- timestamp timestamp,
-- origin_id text,
-- exit_status int
-- );


-- create table initialize_put_events (
-- pipeline text,
-- job text,
-- build text,
-- timestamp timestamp,
-- origin_id text
-- );
-- create table start_put_events (
-- pipeline text,
-- job text,
-- build text,
-- timestamp timestamp,
-- origin_id text
-- );
-- create table finish_put_events (
-- pipeline text,
-- job text,
-- build text,
-- timestamp timestamp,
-- origin_id text,
-- exit_status int
-- );

-- create table initialize_get_events (
-- pipeline text,
-- job text,
-- build text,
-- timestamp timestamp,
-- origin_id text
-- );
-- create table start_get_events (
-- pipeline text,
-- job text,
-- build text,
-- timestamp timestamp,
-- origin_id text
-- );
-- create table finish_get_events (
-- pipeline text,
-- job text,
-- build text,
-- timestamp timestamp,
-- origin_id text,
-- exit_status int
-- );
