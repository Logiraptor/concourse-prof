import React, {useState, useEffect} from 'react';
import { Link } from "react-router-dom";
import {Panel} from 'pivotal-ui/react/panels';
import {Table} from 'pivotal-ui/react/table';

export const JobList = ({apiClient, pipeline}) => {
    var [jobs, setJobs] = useState([]);
    var [loading, setLoading] = useState(false);

    async function fetchJobs() {
        setLoading(true);
        var jobs = await apiClient.listJobs(pipeline);
        setJobs(jobs);
        setLoading(false);
    }

    useEffect(() => {
        fetchJobs();
    }, [apiClient, pipeline]);

    var jobData = jobs.map(name => ({name, link: <Link to={`/pipeline/${pipeline}/job/${name}`}>{name}</Link>}));

    return (
        <Panel loading={loading} className="paxl" header="Jobs">
          <Table data={jobData}/>
        </Panel>
    );
}
