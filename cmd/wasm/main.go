//go:build js && wasm

package main

import (
	"15-puzzle/internal/model"
	"15-puzzle/internal/puzzle"
	"15-puzzle/internal/web-service/handler"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"syscall/js"
)

func main() {
	if err := puzzle.Init(func(p *puzzle.Controller) {
		p.InfoRequest = func() {
			js.Global().Call("apiRequest", js.ValueOf(http.MethodGet), js.ValueOf("api/info"))
		}
		p.UserStatsRequest = func() {
			js.Global().Call("apiRequest", js.ValueOf(http.MethodGet), js.ValueOf("api/stats"))
		}
		p.OnGameStart = func() {
			js.Global().Call("apiRequest", js.ValueOf(http.MethodPut), js.ValueOf("api/start"))
		}
		p.OnGameSolve = func(moves int) {
			js.Global().Call("apiRequest", js.ValueOf(http.MethodPut), js.ValueOf("api/solve?moves="+strconv.Itoa(moves)))
		}
		p.MonitoringRequest = func(code string) {
			js.Global().Call("apiRequest", js.ValueOf(http.MethodGet), js.ValueOf("api/monitoring"), js.ValueOf(code))
		}
		p.UrlOpener = func(url string) {
			js.Global().Call("openLink", js.ValueOf(url))
		}
		js.Global().Set("wasmHTTPRequest", HTTPRequestFunc(p.ApiResponseHandler))
		js.Global().Set("wasmDebug", js.FuncOf(func(this js.Value, args []js.Value) any { p.Debug(args[0].String()); return nil }))
		js.Global().Set("wasmOnLoad", js.FuncOf(func(this js.Value, args []js.Value) any {
			p.OnLoad(args[0].Float(), args[1].String(), args[2].String())
			return nil
		}))
		js.Global().Set("wasmSetActive", js.FuncOf(func(this js.Value, args []js.Value) any { p.SetActive(args[0].Bool()); return nil }))
	}); err != nil {
		fmt.Fprintf(os.Stderr, "start failed: %v", err)
		os.Exit(1)
	}
}

func HTTPRequestFunc(responseHandler func(model.ApiResponse)) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		// Get the URL as argument
		reqMethod := args[0].String()
		reqUrl := args[1].String()
		code := args[2].String()
		appData := args[3].String()
		handler := js.FuncOf(func(this js.Value, args []js.Value) any {
			resolve := args[0]
			reject := args[1]
			onErr := func(err error) {
				errorConstructor := js.Global().Get("Error")
				errorObject := errorConstructor.New(err.Error())
				reject.Invoke(errorObject)
				errStr := err.Error()
				responseHandler(model.ApiResponse{Err: &errStr})
			}
			var result model.ApiResponse
			go func() {
				// The HTTP request
				req, err := http.NewRequest(reqMethod, reqUrl, nil)
				if err != nil {
					onErr(fmt.Errorf("http new request: %s", err))
					return
				}
				req.Header.Add(handler.WebAppInitDataHeader, appData)
				if code != "" {
					req.Header.Add(handler.WebAppExtraCodeHeader, code)
				}
				res, err := http.DefaultClient.Do(req)
				if err != nil {
					onErr(fmt.Errorf("http do request: %s", err))
					return
				}
				defer res.Body.Close()

				data := make([]byte, 0)

				// Func to return a Promise because HTTP requests are blocking in Go
				returnPromise := func() {
					arrayConstructor := js.Global().Get("Uint8Array")
					dataJS := arrayConstructor.New(len(data))
					js.CopyBytesToJS(dataJS, data)
					responseConstructor := js.Global().Get("Response")
					response := responseConstructor.New(dataJS)
					resolve.Invoke(response)
				}

				switch res.StatusCode {
				case http.StatusForbidden:
					returnPromise()
					return
				case http.StatusOK:
				default:
					onErr(fmt.Errorf("http code: %d", res.StatusCode))
					return
				}

				// Read the response body
				data, err = io.ReadAll(res.Body)
				if err != nil {
					onErr(fmt.Errorf("http read response: %s", err))
					return
				}
				if err := json.Unmarshal(data, &result); err != nil {
					onErr(fmt.Errorf("http response to json: %s", err))
					return
				}

				responseHandler(result)
				returnPromise()
			}()
			return nil
		})
		promiseConstructor := js.Global().Get("Promise")
		return promiseConstructor.New(handler)
	})
}
