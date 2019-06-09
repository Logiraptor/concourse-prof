import React, {useState, useEffect} from 'react';
import {Panel} from 'pivotal-ui/react/panels';
import Highcharts from 'highcharts';
import HighchartsReact from 'highcharts-react-official';

export const toHHMMSS = (millis) => {
    var total_seconds = millis / 1000;
    var hours         = Math.floor(total_seconds / 3600);
    var minutes       = Math.floor((total_seconds - (hours * 3600)) / 60);
    var seconds       = total_seconds - (hours * 3600) - (minutes * 60);

    if (hours   < 10) {hours   = "0"+hours;}
    if (minutes < 10) {minutes = "0"+minutes;}
    if (seconds < 10) {seconds = "0"+seconds;}
    return hours + ':' + minutes + ':' + seconds;
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
            xx_duration: start - init,
        });
        data.push({
            x: start,
            x2: finish,
            partialFill: 1,
            y: i,
            xx_duration: finish - start,
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
            borderColor: 'gray',
            pointWidth: 20,
            data: data,
            dataLabels: {
                formatter: function() {
                    return toHHMMSS(this.point.xx_duration)
                },
                enabled: true
            }
        }]
    };
};

export const BuildPlot = ({plotter, apiClient, pipeline, job, build}) => {
    var [plot, setPlot] = useState([]);

    useEffect(() => {
        plotter.plotBuild(pipeline, job, build).subscribe(plot => {
            setPlot(plot);
        });
    }, [pipeline, job, build]);

    return (
        <Panel className="paxl" header="Plot">
             <HighchartsReact
                highcharts={Highcharts}
                options={makeOptions(plot)} />
        </Panel>
    );
}
