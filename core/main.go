package main

import (
	"fmt"
	"log"
	"sync"
	"syscall/js"
	"time"

	"github.com/Logiraptor/concourse-prof/processor"
)

func main() {

	var jsDate = js.Global().Get("Date")
	var makeDate = func(x time.Time) js.Value {
		return jsDate.New(
			js.ValueOf(x.Year()),
			js.ValueOf(int(x.Month()-1)),
			js.ValueOf(x.Day()),
			js.ValueOf(x.Hour()),
			js.ValueOf(x.Minute()),
			js.ValueOf(x.Second()),
		)
	}

	js.Global().Set("NewProcessor", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		concourseUrl := args[0]
		localUrl := args[1]
		token := args[2]

		client, err := NewApiClient(concourseUrl.String(), localUrl.String(), token.String())
		if err != nil {
			log.Println(err)
			return js.Null()
		}

		return map[string]interface{}{
			"listPipelines": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				cb := args[0]
				go func() {
					pipelines, err := client.ListPipelines()
					if err != nil {
						fmt.Println(err)
						return
					}
					output := []interface{}{}
					for _, pipeline := range pipelines {
						output = append(output, pipeline)
					}
					cb.Invoke(output)
				}()
				return js.Null()
			}),
			"listJobs": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				pipeline := args[0]
				cb := args[1]
				go func() {
					jobs, err := client.ListJobs(pipeline.String())
					if err != nil {
						fmt.Println(err)
						return
					}
					output := []interface{}{}
					for _, job := range jobs {
						output = append(output, job)
					}
					cb.Invoke(output)
				}()
				return js.Null()
			}),
			"listBuilds": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				pipeline := args[0]
				job := args[1]
				cb := args[2]
				go func() {
					builds, err := client.ListBuilds(pipeline.String(), job.String())
					if err != nil {
						fmt.Println(err)
						return
					}
					output := []interface{}{}
					for _, build := range builds {
						output = append(output, build)
					}
					cb.Invoke(output)
				}()
				return js.Null()
			}),
			"plotBuild": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				pipeline := args[0]
				job := args[1]
				build := args[2]
				cb := args[3]
				go func() {
					wg := &sync.WaitGroup{}
					sink := &plotterEventSink{}
					ui := consoleUi{}
					p := processor.NewProcessor(client, ui, sink, wg)
					p.ProcessBuild(pipeline.String(), job.String(), build.String())
					p.Close()
					wg.Wait()
					output := []interface{}{}
					for origin, interval := range sink.Intervals {
						fmt.Println(origin)
						fmt.Println(interval.Init, interval.Start, interval.Finish)
						output = append(output, map[string]interface{}{
							"origin": origin,
							"init":   makeDate(interval.Init),
							"start":  makeDate(interval.Start),
							"finish": makeDate(interval.Finish),
							"today":  makeDate(time.Now()),
						})
					}
					cb.Invoke(output)
				}()
				return js.Null()
			}),
		}
	}))

	select {}
}
