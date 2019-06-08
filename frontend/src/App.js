import React, {useState, useEffect} from 'react';
import { HashRouter as Router, Route } from "react-router-dom";
import './App.css';
import {Siteframe} from 'pivotal-ui/react/siteframe';
import {Input} from 'pivotal-ui/react/inputs';
import {PrimaryButton} from 'pivotal-ui/react/buttons';
import {Panel} from 'pivotal-ui/react/panels';
import {Grid, FlexCol} from 'pivotal-ui/react/flex-grids';
import {Icon} from 'pivotal-ui/react/iconography';
import 'pivotal-ui/css/whitespace';
import {ApiClient} from './core/ApiClient';
import {Plotter} from './core/plotter';
import {PipelineList} from './PipelineList';
import {JobList} from './JobList';
import {BuildList} from './BuildList';
import {BuildPlot} from './BuildPlot';

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
};

export default App;
