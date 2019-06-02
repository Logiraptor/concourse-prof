import React, {useState, useEffect} from 'react';
import { HashRouter as Router, Route, Link } from "react-router-dom";
import './App.css';
import {Siteframe} from 'pivotal-ui/react/siteframe';
import {Input} from 'pivotal-ui/react/inputs';
import {PrimaryButton} from 'pivotal-ui/react/buttons';
import {Panel} from 'pivotal-ui/react/panels';
import {Grid, FlexCol} from 'pivotal-ui/react/flex-grids';
import {Table} from 'pivotal-ui/react/table';
import {Icon} from 'pivotal-ui/react/iconography';
import 'pivotal-ui/css/whitespace';
import Highcharts from 'highcharts';
import HC_xrange from 'highcharts/modules/xrange';
import HC_exporting from 'highcharts/modules/exporting';
import HighchartsReact from 'highcharts-react-official';
import {ApiClient} from './core/api_client';
import {Plotter} from './core/plotter';

HC_xrange(Highcharts);
HC_exporting(Highcharts);

const useStateWithLocalStorage = (localStorageKey, defaultVal) => {
    const [value, setValue] = useState(
        localStorage.getItem(localStorageKey) || defaultVal
    );

    useEffect(() => {
        localStorage.setItem(localStorageKey, value);
    }, [value, localStorageKey]);

    return [value, setValue];
};

const App = () => {

    var [token, setToken] = useStateWithLocalStorage("concourse-api-token", "");
    var [url, setUrl] = useStateWithLocalStorage("concourse-api-url", "");

    var [tempToken, setTempToken] = useState(token);
    var [tempUrl, setTempUrl] = useState(url);

    var apiClient = new ApiClient(url, process.env.PUBLIC_URL, token);
    var plotter = new Plotter(apiClient);

    return (
            <Siteframe
                headerProps={{
                    logo: <div className="ptl pbl pll" style={{fontSize: '40px'}}><Icon src="pivotal_ui_inverted" style={{fill: 'currentColor'}}/></div>,
                    companyName: 'Experimental',
                    productName: 'Concourse Monitor'
                }}
            >
            <div className="paxl" style={{overflow: 'auto', height: '100%'}}>
                <Panel className="paxl" header="Target"
                        headerCols={[<FlexCol fixed><PrimaryButton small onClick={() => {setUrl(tempUrl); setToken(tempToken)}}> Refresh </PrimaryButton></FlexCol>]} >
                    <Grid>
                        <FlexCol>
                            <label>
                                Url <Input value={tempUrl} onChange={event => setTempUrl(event.target.value)}/>
                            </label>
                        </FlexCol>
                        <FlexCol>
                            <label>
                                Token <Input value={tempToken} onChange={event => setTempToken(event.target.value)}/>
                            </label>
                        </FlexCol>
                    </Grid>
                </Panel>

                <Router>
                    <Route exact path="/" render={() => <PipelineList apiClient={apiClient}/>}/>
                    <Route exact path="/pipeline/:pipeline" render={({match}) => <JobList apiClient={apiClient} pipeline={match.params.pipeline}/>}/>
                    <Route exact path="/pipeline/:pipeline/job/:job" render={({match}) => <BuildList apiClient={apiClient} pipeline={match.params.pipeline} job={match.params.job}/>}/>
            <Route exact path="/pipeline/:pipeline/job/:job/build/:build" render={({match}) => <BuildPlot plotter={plotter} apiClient={apiClient} pipeline={match.params.pipeline} job={match.params.job} build={match.params.build}/>}/>
                </Router>
            </div>
            </Siteframe>
    );
}

const PipelineList = ({apiClient}) => {
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

const JobList = ({apiClient, pipeline}) => {
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

const BuildList = ({apiClient, pipeline, job}) => {
    var [builds, setBuilds] = useState([]);
    var [loading, setLoading] = useState(false);

    async function fetchBuilds() {
        setLoading(true);
        var builds = await apiClient.listBuilds(pipeline, job);
        setBuilds(builds);
        setLoading(false);
    }

    useEffect(() => {
        fetchBuilds();
    }, [apiClient, pipeline, job]);

    var buildData = builds.map(name => ({name, link: <Link to={`/pipeline/${pipeline}/job/${job}/build/${name}`}>{name}</Link>}));

    return (
        <Panel loading={loading} className="paxl" header="Builds">
          <Table data={buildData}/>
        </Panel>
    );
}

const makeOptions = (plot) => {
    plot = plot.sort((x, y) => x.start.getTime() - y.start.getTime());
    var categories = plot.map(x => x.origin);
    var data = [];
    plot.forEach((step, i) => {
        var init = step.init.getTime();
        var start = step.start.getTime();
        var finish = step.finish.getTime();
        data.push({
            x: init,
            x2: start,
            partialFill: 1,
            y: i,
        });
        data.push({
            x: start,
            x2: finish,
            partialFill: 1,
            y: i,
        });
    });

    return {
        chart: {
            type: 'xrange'
        },
        title: {
            text: 'Steps'
        },
        xAxis: {
            type: 'datetime'
        },
        yAxis: {
            title: {
                text: 'Step'
            },
            categories: categories,
            reversed: true
        },
        series: [{
            name: 'Job',
            // pointPadding: 0,
            // groupPadding: 0,
            borderColor: 'gray',
            pointWidth: 20,
            data: data,
            dataLabels: {
                enabled: true
            }
        }]
    };
};

const BuildPlot = ({plotter, apiClient, pipeline, job, build}) => {
    var [plot, setPlot] = useState([]);
    var [loading, setLoading] = useState(false);

    useEffect(() => {
        plotter.plotBuild(pipeline, job, build).subscribe(plot => {
            setPlot(plot);
        });
    }, [pipeline, job, build]);

    return (
        <Panel loading={loading} className="paxl" header="Plot">
          {loading ? null :
            <HighchartsReact
                highcharts={Highcharts}
                options={makeOptions(plot)} />}
        </Panel>
    );
}

export default App;
