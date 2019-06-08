import React from 'react';
import ReactDOM from 'react-dom';
import './index.css';
import App from './App';
import * as serviceWorker from './serviceWorker';

import Highcharts from 'highcharts';
import HC_xrange from 'highcharts/modules/xrange';
import HC_exporting from 'highcharts/modules/exporting';

HC_xrange(Highcharts);
HC_exporting(Highcharts);

ReactDOM.render(<App />, document.getElementById('root'));

// If you want your app to work offline and load faster, you can change
// unregister() to register() below. Note this comes with some pitfalls.
// Learn more about service workers: https://bit.ly/CRA-PWA
serviceWorker.unregister();
