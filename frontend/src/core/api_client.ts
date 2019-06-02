import { Observable, Observer } from 'rxjs';
import { Event } from './events';
import { EventSourcePolyfill } from 'event-source-polyfill';


export class ApiClient {
    constructor(
        private url: string,
        private localUrl: string,
        private token: string,
    ) {}

    request = async (path: string) => {
        const data = await fetch(`${this.localUrl}${path}`, {
            headers: {
                "X-Concourse-Url": this.url,
                "Authorization": this.token,
            },
        });
        return await data.json();
    }

    async listPipelines(): Promise<string[]> {
        const pipelines = await this.request(`/api/v1/teams/main/pipelines`);
        return pipelines.map((x: {name: string}) => x.name)
    }

    async listJobs(pipeline: string): Promise<string[]> {
        const jobs = await this.request(`/api/v1/teams/main/pipelines/${pipeline}/jobs`);
        return jobs.map((x: {name: string}) => x.name)
    }

    async listBuilds(pipeline: string, job: string): Promise<number> {
        const builds = await this.request(`/api/v1/teams/main/pipelines/${pipeline}/jobs/${job}/builds?limit=50`);
        return builds.map((x: {id: number}) => x.id)
    }

    async getPlan(build: string): Promise<unknown> {
        return await this.request(`/api/v1/builds/${build}/plan`);
    }

    listEvents(pipeline: string, job: string, build: string): Observable<Event> {
        return Observable.create((observer: Observer<Event>) => {
            const es = new EventSourcePolyfill(`/api/v1/builds/${build}/events`, {
                headers: {
                    "X-Concourse-Url": this.url,
                    "Authorization": this.token,
                },
            });

            es.addEventListener("event", event => {
                observer.next(JSON.parse((event as MessageEvent).data) as Event);
            });

            es.addEventListener("end", (event) => {
                es.close();
                observer.complete();
            });

            es.onerror = (event) => {
                observer.error(event);
            };
        });
    }
}

