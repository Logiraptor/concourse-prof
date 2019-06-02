
export type Event =
    | {event:  "initialize-task", data: InitializeTaskEvent}
    | {event:  "start-task", data: StartTaskEvent}
    | {event:  "finish-task", data: FinishTaskEvent}
    | {event:  "initialize-put", data: InitializePutEvent}
    | {event:  "start-put", data: StartPutEvent}
    | {event:  "finish-put", data: FinishPutEvent}
    | {event:  "initialize-get", data: InitializeGetEvent}
    | {event:  "start-get", data: StartGetEvent}
    | {event:  "finish-get", data: FinishGetEvent}
    | {event:  "log", data: LogEvent}
    | {event:  "status", data: StatusEvent}
    | {event:  "error", data: ErrorEvent}

export interface Origin {
    id: string;
}

export interface StatusEvent {
    time: number;
    status: string;
}

export interface LogEvent {
    time: number;
    payload: string;
    origin: Origin;
}

export interface ErrorEvent {
    message: string;
}

export interface InitializeGetEvent {
    time: number;
    origin: Origin;
}

export interface StartGetEvent {
    time: number;
    origin: Origin;
}

export interface FinishGetEvent {
    time: number;
    exitStatus: number;
    origin: Origin;
    version: {[x: string]: string};
    metadata: {name: string, value: string}[];
}

export interface InitializePutEvent {
    time: number;
    origin: Origin;
}

export interface StartPutEvent {
    time: number;
    origin: Origin;
}

export interface FinishPutEvent {
    time: number;
    exitStatus: number;
    origin: Origin;
    version: {[x: string]: string};
    metadata: Array<{name: string, value: string}>;
}

export interface InitializeTaskEvent {
    time: number;
    origin: Origin;
    config: {
        platform: string;
        image: string;
        run: {
            path: string;
            args: string[];
            dir: string;
        }
        inputs: Array<{name: string, value: string}>;
    }
}

export interface StartTaskEvent {
    time: number;
    origin: Origin;
    config: {
        platform: string;
        image: string;
        run: {
            path: string;
            args: string[];
            dir: string;
        }
        inputs: Array<{name: string, value: string}>;
    }
}

export interface FinishTaskEvent {
    time: number;
    exitStatus: number;
    origin: Origin;
}
