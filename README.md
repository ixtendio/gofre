<h1 align="center">GOFre - A Sweet Web Framework for Go</h1>

<p align="center">
<img alt="GOFre" src="docs/img/logo.png" />
</p>

_GOFree[^1]_ if a web framework for Go, without third-party party dependencies, that makes the development of the web applications a joy.  _GOFre_ integrates with `http.Server` and supports the standard Go HTTP handlers: `http.Handler` and `http.HandlerFunc`. 

This framework was developed around simplicity of usage and extensibility and offers the following features: 
* **Middleware**
* **Path pattern matching** - including path variable extraction and validation
* **Templating** - including static resources
* **Authentication** - OAUTH2 flow included for GitHub and Google
* **Authorization** - RBAC implementation
* **SSE (Server Sent-Events)**
* **Security** - CSRF Middleware protection

## Architecture Overview

_GOFre_ has the following components:
* **HttpRequest** - an object that encapsulates the initial `http.Request` and the path variables, if exists 
* **HttpResponse** - an object that encapsulates the response and knows how to write it back to the client
* **Handler** - a function that receives a `Context` and an `HttpRequest` and returns an `HttpResponse` or an `error`
* **Middleware** - a function that receives a `Handler` and returns another `Handler`
* **Router** - an object that knows how to parse the `http.Request` and to route it to the corresponding `Handler`

![Architecture](docs/img/gofre-architecture.png)

### Path Pattern Matching

_GOFre_ supports a complex path matching where the most specific pattern is chosen first. 

Supported path patterns matching:

1. **exact matching**  - `/a/b/c`
2. **capture variable without constraints** 
   1. `/a/{b}/{c}`
      1. `/a/john/doe` => b: john, c: doe
3. **capture variable with constraints** 
   1. `/a/{uuid:^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$}` - UUID matching
      1. `/a/zyw3040f-0f1c-4e98-b71c-d3cd61213f90` => false (z,x and w are not part of UUID regex)
      2. `/a/fbd3040f-0f1c-4e98-b71c-d3cd61213f90` => true
   2. `/a/{number:^[0-9]{3}$}` - number with 3 digits
      1. `/a/12` => false
      1. `/a/123` => true
      1. `/a/012` => true
      1. `/a/0124` => false
4. **literal match regex** 
   1. **&ast;** - matches any number of characters or a single segment path
      1. `/a/abc*hij`
         1. `/a/abcdhij` => true
         2. `/a/abcdefghij` => true
         3`/a/abcdefgij` => false (the path doesn't end with `hij`)
      2. `/a/*/c`
         1. `/a/b/c` => true
         2. `/a/b/c/c` => false (max 3 path segments allowed, and we have 4)
      3. `/a/abc*hij/*`
         1. `/a/abcdefghij/abc` => true
         2. `/a/abcdefghij/abc/xyz` => false (`*` matches a single path segment and, we have 2 `abc/xyz`)
   2. **?** - matches a single character
      1. `/a/abc?hij`
         1. `/a/abcdhij` => true
         2. `/a/abcdehij` => false (the character `e` will not match)
5. **greedy match** 
   1. **&ast;&ast;** - matches multiple path segments
      1. `/a/**/z`
         1. `/a/b/c/d/e/f/z` => true
         1. `/a/b/c/d/e/f` => false (the path should end in `/z`)
      2. `/a/**`
         1. `/a/b/c/d/e/f` => true

Comparing with other libraries, _GOFre_ does not require to declare the path patterns in a specific order so that the match to work as you expect. 

For example, these path matching patterns (assuming we handle only GET requests) can be declared in any order in your code: 

1. `/users/john/{lastName}`
2. `/users/john/doe`
3. `/users/*/doe`

Here are some URL's example with their matched pattern:

* `https://www.website.com/users/john/doe` - the second pattern will match
* `https://www.website.com/users/john/wick` - the first pattern will match, where the lastName will be `wick`
* `https://www.website.com/users/jane/doe` - the third pattern will match

_GOFre_ includes also support for greedy path matching: `**`

* `/users/**/doe` - matches any path that starts with `/users` and ends with `/doe` 
* `/users/**` - matches any path that starts with `/users`

The path matching can be **case-sensitive** (default) or **case-insensitive**.

If two path patterns (of the same type) that matches the same URL are registered, then the framework will panic. For example: 
* `/a/{b}`
* `/a/{d}`

On the other side, the following two patterns are accepted by the framework, although the second one will never be executed (this is a limitation of the path matching that might be solved in the future releases)
* `/a/{b}`
* `/a/*`

### Middlewares

A _middleware_ is a function that intercepts a request. The function receives as an argument a _Handler_ and returns another _Handler_

There are two ways to register the middlewares:
* **common registration** - applied to all handlers. A common middleware can be:
  * **pre registered** - executed before the custom handler middlewares
  * **post registered** - executed after the custom handler middlewares
* **per handler registration** - applied for a single handler only

Example:

```go 
gowMux.CommonPreMiddlewares(func(handler handler.Handler) handler.Handler {
  return func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
      log.Println("Common pre middleware 1 - before processing the request")
      resp, err := handler(ctx, r)
      log.Println("Common pre middleware 1 - after processing the request")
      return resp, err
  }
}, func(handler handler.Handler) handler.Handler {
  return func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
      log.Println("Common pre middleware 2 - before processing the request")
      resp, err := handler(ctx, r)
      log.Println("Common pre middleware 2 - after processing the request")
      return resp, err
  }
})

gowMux.CommonPostMiddlewares(func(handler handler.Handler) handler.Handler {
  return func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
      log.Println("Common post middleware 1 - before processing the request")
      resp, err := handler(ctx, r)
      log.Println("Common post middleware 1 - after processing the request")
      return resp, err
  }
}, func(handler handler.Handler) handler.Handler {
  return func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
      log.Println("Common post middleware 2 - before processing the request")
      resp, err := handler(ctx, r)
      log.Println("Common post middleware 2 - after processing the request")
      return resp, err
  }
})

gowMux.HandleGet("/handlers", func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
  log.Println("Request handling")
  return response.PlainTextHttpResponseOK("ok"), nil
}, func(handler handler.Handler) handler.Handler {
  return func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
      log.Println("Custom middleware 1 - before processing the request")
      resp, err := handler(ctx, r)
      log.Println("Custom middleware 1 - after processing the request")
      return resp, err
  }
}, func(handler handler.Handler) handler.Handler {
  return func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
      log.Println("Custom middleware 2 - before processing the request")
      resp, err := handler(ctx, r)
      log.Println("Custom middleware 2 - after processing the request")
      return resp, err
  }
})
```

If we execute `curl -vvv "https://localhost:8080/handlers"` we should see the following lines in the console:

```text
Common pre middleware 1 - before processing the request
Common pre middleware 2 - before processing the request
Custom middleware 1 - before processing the request
Custom middleware 2 - before processing the request
Common post middleware 1 - before processing the request
Common post middleware 2 - before processing the request
Request handling
Common post middleware 2 - after processing the request
Common post middleware 1 - after processing the request
Custom middleware 2 - after processing the request
Custom middleware 1 - after processing the request
Common pre middleware 2 - after processing the request
Common pre middleware 1 - after processing the request
```

The _middleware_ package includes the following middlewares:

* **Panic** - handles the panic and converts them to an error
* **ErrResponse** - converts an error to an HTTP answer
* **CSRFPrevention** - provides basic CSRF protection for a web application
* **Cors** - enable client-side cross-origin requests by implementing W3C's CORS
* **Authorize** - provides basic RBAC authorization (authentication is required in this case)



## Installation

You can install this repo with `go get`:
```sh
go get github.com/ixtendio/gofre
```
## Usage

```go
gowConfig := &gofre.Config{
	CaseInsensitivePathMatch: false,
	ContextPath:              "",
	ErrLogFunc: func(err error) {
		log.Printf("An error occurred in the gofre framework: %v", err)
	},
}
gowMux, err := gofre.NewMuxHandler(gowConfig)
if err != nil {
	log.Fatalf("Failed to create gofre mux handler, err: %v", err)
}

// JSON with vars path
gowMux.HandleGet("/hello/{firstName}/{lastName}", func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
	return response.JsonHttpResponseOK(r.UriVars), nil
})

httpServer := http.Server{
	Addr:              ":8080",
	Handler:           gowMux,
}
if err := httpServer.ListenAndServe(); err != nil {
	log.Fatalf("Failed starting the HTTP server, err: %v", err)
}
```

```shell
curl -vvv "https://localhost:8080/hello/John/Doe"
```

# Run the examples

 1. Execute the make file:
    1. MacOS `make run-osx`
    2. Linux `make run`
 2. In the browser, open the following URL: `https://locahost:8080`  

[^1]: Gofri (singular **gofre**) are waffles in Italy and can be found in the Piedmontese cuisine: they are light and crispy in texture.
