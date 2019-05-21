
with

-- split all the log_events on \n to produce one row per line
expanded_log_events as (
  select pipeline, job, build, origin, time, ARRAY[event_id::int, x.idx::int] as event_id, x.payload as payload
  from log_events,
       unnest(regexp_split_to_array(payload, E'\n')) with ordinality as x(payload, idx)
  where x.payload <> ''
  order by origin, event_id, x.idx
),

-- increment a counter on each row which 'looks like' a bash set -x command ie "+ popd"
cmd_groups as (
  select pipeline, job, build, origin, payload, time, event_id,
    count(case payload ~ '\++ [a-z0-9]+ .*' when true then 1 else null end) over (partition by pipeline, job, build, origin order by event_id) as cmd_group_index
  from expanded_log_events
),

-- concat all logs which come after the command
grouped_logs as (
  select pipeline, job, build, origin, string_agg(payload, '\n') as payload, min(time) as time, min(event_id) as event_id
  from cmd_groups
  group by pipeline, job, build, origin, cmd_group_index
  order by event_id
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

-- time each command by looking ahead to the timestamp of the next command
-- For the last command in the task, we use the finish event to determine duration
-- Pre-concourse-5, get and put events did not record finish times, so those are left with NULL durations
with_durations as (
  select gl.*,
  coalesce(LEAD(gl.time) OVER (partition by gl.pipeline, gl.job, gl.build, gl.origin ORDER BY gl.event_id), case f.time = 0 when true then null else f.time end) - gl.time as duration
  from grouped_logs gl
  inner join finish_events f on f.build = gl.build and gl.origin = f.origin
)

-- grab top commands by duration
select pipeline, job, build, origin, duration, left(payload, 75) as payload
from with_durations
order by duration desc nulls last;

