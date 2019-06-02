import { ApiClient } from './api_client';
import { Observable, from } from 'rxjs';
import { reduce, map, concatMap } from 'rxjs/operators';

interface Interval {
    init: Date;
    start: Date;
    finish: Date;
}

type JobPlot = Interval[]

export class Plotter {
    constructor(private apiClient: ApiClient) {}

    plotBuild(pipeline: string, job: string, build: string): Observable<JobPlot> {

        return from(this.apiClient.getPlan(build)).pipe(concatMap(plan => {

            const idLookup = new Map<string, string>();
            function traverse(plan: unknown) {
                if (plan instanceof Array) {
                    plan.forEach(traverse);
                } else if (plan instanceof Object) {
                    if ("id" in plan) {
                        let typ: (keyof typeof plan) | undefined;
                        for (const key of Object.keys(plan)) {
                            if (key !== "id") {
                                typ = key;
                                break;
                            }
                        }

                        let name = "";
                        if (typ && "name" in plan[typ]) {
                            name = (plan[typ] as any).name;
                        }

                        idLookup.set((plan as any).id, `${name}/${String(typ)}`);
                    }
                    Object.values(plan).forEach(traverse);
                }
            }
            traverse(plan);

            return this.apiClient.listEvents(pipeline, job, build).pipe(
                reduce((plot, event) => {
                    if ("origin" in event.data) {
                        const originName = idLookup.get(event.data.origin.id) || '???';
                        const interval: Interval = plot.get(originName) || {
                            init: new Date(), start: new Date(), finish: new Date(),
                        };
                        switch (event.event) {
                            case "initialize-get":
                            case "initialize-put":
                            case "initialize-task":
                                interval.init = new Date(event.data.time * 1000);
                                break;

                            case "start-get":
                            case "start-put":
                            case "start-task":
                                interval.start = new Date(event.data.time * 1000);
                                break;

                            case "finish-get":
                            case "finish-put":
                            case "finish-task":
                                interval.finish = new Date(event.data.time * 1000);
                                break;
                        }
                        plot.set(originName, interval);
                    }
                    return plot
                }, new Map<string, Interval>()),
                map(intervalMap => {
                    const result: Array<{origin: string} & Interval> = [];
                    intervalMap.forEach((interval, origin) => {
                        result.push({
                            origin,
                            ...interval,
                        });
                    });
                    return result;
                }),
            );
        }));
    }
}
