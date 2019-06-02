
declare module 'event-source-polyfill' {
    interface EventSourcePolyfillInit {
        withCredentials?: boolean;
        headers: {[key: string]: string};
    };

    var EventSourcePolyfill : typeof EventSource & {
        new(url: string, eventSourceInitDict?: EventSourcePolyfillInit): EventSource;
    };

}
