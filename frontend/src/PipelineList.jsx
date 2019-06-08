import React, {useState, useEffect} from 'react';
import { Link } from "react-router-dom";
import {Panel} from 'pivotal-ui/react/panels';
import {Table} from 'pivotal-ui/react/table';

export const PipelineList = ({apiClient}) => {
    var [pipelines, setPipelines] = useState([]);
    var [loading, setLoading] = useState(false);

    async function fetchPipelines() {
        setLoading(true);
        var pipelines = await apiClient.listPipelines();
        setPipelines(pipelines);
        setLoading(false);
    }

    useEffect(() => {
        fetchPipelines();
    }, [apiClient]);

    var pipelineData = pipelines.map(name => ({name, link: <Link to={`/pipeline/${name}`}>{name}</Link>}));

    return (
        <Panel loading={loading} className="paxl" header="Pipelines">
          <Table data={pipelineData}/>
        </Panel>
    );
}
