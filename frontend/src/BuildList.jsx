import React, {useState, useEffect} from 'react';
import { Link } from "react-router-dom";
import {Panel} from 'pivotal-ui/react/panels';
import {Table} from 'pivotal-ui/react/table';
import Highcharts from 'highcharts';
import HighchartsReact from 'highcharts-react-official';

const makeOptions = (successfulBuilds, failedBuilds) => {
    const successData = successfulBuilds.map(build => {
        return {
            x: build.startTime,
            y: build.duration,
        }
    });
    const failureData = failedBuilds.map(build => {
        return {
            x: build.startTime,
            y: build.duration,
        }
    });

    return {
        title: {
            text: 'Job Duration'
        },
        xAxis: {
            type: 'datetime'
        },
        yAxis: {
            title: {
                text: 'Duration'
            },
        },
        series: [{
            name: 'Successful Builds',
            borderColor: 'gray',
            pointWidth: 20,
            data: successData,
            dataLabels: {
                enabled: true
            }
        }, {
            name: 'Failed Builds',
            borderColor: 'gray',
            pointWidth: 20,
            data: failureData,
            dataLabels: {
                enabled: true
            }
        }]
    };
};

export const BuildList = ({apiClient, pipeline, job}) => {
    const [builds, setBuilds] = useState([]);
    const [loading, setLoading] = useState(false);

    async function fetchBuilds() {
        setLoading(true);
        const builds = await apiClient.listBuilds(pipeline, job);
        setBuilds(builds);
        setLoading(false);
    }

    useEffect(() => {
        fetchBuilds();
    }, [apiClient, pipeline, job]);

    const buildData = builds.map(build => ({
        ...build,
        link: (
            <Link to={`/pipeline/${pipeline}/job/${job}/build/${build.id}`}>
                {build.name}
            </Link>
        ),
    }));

    const successes = buildData.filter(x => x.status === 'succeeded');
    const failures = buildData.filter(x => x.status === 'failed');
    const totalNonErrored = successes.length + failures.length;
    const successRate = ((successes.length / totalNonErrored) * 100).toFixed(2);

    return (
        <>
            <Panel loading={loading} className="paxl" header="Overview">
              <div className="em-max aligner txt-c">
                <h1>{successRate}%  Success Rate</h1>
                ({successes.length} / {totalNonErrored})
                <HighchartsReact
                  highcharts={Highcharts}
                  options={makeOptions(successes, failures)} />
              </div>
            </Panel>
            <Panel loading={loading} className="paxl" header="Builds">
                <Table data={buildData}/>
            </Panel>
        </>
    );
}
