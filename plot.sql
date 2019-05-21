
with

initialize_events as (
select pipeline, job, build, origin, time
from initialize_task_events
UNION ALL
select pipeline, job, build, origin, time
from initialize_get_events
UNION ALL
select pipeline, job, build, origin, time
from initialize_put_events
),

start_events as (
select pipeline, job, build, origin, time
from start_task_events
UNION ALL
select pipeline, job, build, origin, time
from start_get_events
UNION ALL
select pipeline, job, build, origin, time
from start_put_events
),

finish_events as (
  select pipeline, job, build, origin, time
  from finish_task_events
  UNION ALL
  select pipeline, job, build, origin, time
  from finish_get_events
  UNION ALL
  select pipeline, job, build, origin, time
  from finish_put_events
),

steps as (
  select
    i.pipeline, i.job, i.build, i.origin,
    i.time as init_time,
    s.time as start_time,
    f.time as finish_time,
    s.time - i.time as init_duration,
    f.time - s.time as run_duration,

    st2.time as job_finish_time,
    st1.time as job_start_time,

    st2.time - st1.time as job_duration
  from initialize_events i
  inner join start_events s on i.origin = s.origin and i.build = s.build
  inner join finish_events f on i.origin = f.origin and i.build = f.build
  inner join status_events st1 on i.build = st1.build
  inner join status_events st2 on i.build = st2.build and st2.event_id > st1.event_id
),

with_plot_points as (
   select pipeline, job, build, origin,
     ((init_time - job_start_time)::float / (job_duration::float)) * 100 as init_offset_pct,
     ((start_time - init_time)::float / (job_duration::float)) * 100 as start_offset_pct,
     ((finish_time - start_time)::float / (job_duration::float)) * 100 as finish_offset_pct
   from steps
   where job_duration <> 0
),

with_plot as (
   select pipeline, job, build, origin,
     rpad(concat(
       repeat(' ', init_offset_pct::int),
       repeat('-', start_offset_pct::int),
       repeat('X', finish_offset_pct::int)), 100, ' ') as timeline
   from with_plot_points
   order by pipeline, job, build, init_offset_pct, origin
)

select * from with_plot;
