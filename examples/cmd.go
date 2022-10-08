package main

import (
	"context"
	"github.com/ixtendio/gow"
	"github.com/ixtendio/gow/auth"
	"github.com/ixtendio/gow/auth/oauth"
	"github.com/ixtendio/gow/cache"
	"github.com/ixtendio/gow/middleware"
	"github.com/ixtendio/gow/request"
	"github.com/ixtendio/gow/response"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {

	ctx := context.Background()
	gowConfig := &gow.Config{
		CaseInsensitivePathMatch: false,
		ContextPath:              "",
		ResourcesConfig: &gow.ResourcesConfig{
			TemplatesPathPattern: "examples/resources/templates/*.html",
			AssetsDirPath:        "./examples/resources/assets",
		},
		ErrLogFunc: func(err error) {
			log.Printf("An error occurred: %v", err)
		},
	}
	gowMux, err := gow.NewMuxHandler(gowConfig)
	if err != nil {
		log.Fatalf("Failed to create gow handler, err: %v", err)
	}
	gowMux.RegisterCommonMiddlewares(middleware.Panic())

	// template example
	gowMux.HandleGet("/", func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
		return response.TemplateHttpResponseOK(gowConfig.ResourcesConfig.Template, "index.html", nil), nil
	})

	// OAUTH2 flow with user details extraction
	gowMux.HandleOAUTH2(oauth.Config{
		WebsiteUrl:       "https://localhost:8080",
		FetchUserDetails: true,
		Providers: []oauth.Provider{
			oauth.GitHubProvider{
				ClientId:     os.Getenv("GITHUB_OAUTH_CLIENT_ID"),
				ClientSecret: os.Getenv("GITHUB_OAUTH_CLIENT_SECRET"),
			},
			oauth.GoogleProvider{
				ClientId:     os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
				ClientSecret: os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
				Scopes:       []string{"openid"},
			}},
		CacheConfig: oauth.CacheConfig{
			Cache:             cache.NewInMemory(),
			KeyExpirationTime: 1 * time.Minute,
		},
	}, func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
		accessToken := oauth.GetAccessTokenFromContext(ctx)
		securityPrincipal := auth.GetSecurityPrincipalFromContext(ctx)
		return response.JsonHttpResponseOK(map[string]interface{}{
			"accessToken":       accessToken,
			"authenticatedUser": securityPrincipal,
		}), nil
	})

	// TEXT plain response
	gowMux.HandleGet("/text/plain", func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
		return response.PlainTextHttpResponseOK("Text plain response"), nil
	})

	// HTML response
	gowMux.HandleGet("/text/html", func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
		return response.HtmlHttpResponseOK("<!DOCTYPE html><html><body><h1>HTML example</h1></body></html>"), nil
	})

	// JSON with vars path
	gowMux.HandleGet("/json/{user}/{id}", func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
		return response.JsonHttpResponseOK(r.UriVars), nil
	})

	// document download example
	gowMux.HandleGet("/download", func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
		f, err := os.Open("./examples/resources/assets/image.png")
		if err != nil {
			return nil, err
		}
		return response.StreamHttpResponse(f, "image/png"), nil
	})

	// template example
	gowMux.HandleGet("/tmpl/{tmplName}", func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
		templateName := r.UriVars["tmplName"] + ".html"
		return response.TemplateHttpResponseOK(gowConfig.ResourcesConfig.Template, templateName, nil), nil
	})

	// SSE example
	evtGen := NewEventGenerator(ctx, 10)
	gowMux.HandleGet("/sse", func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
		return response.SSEHttpResponse(func(ctx context.Context, lastEventId string) <-chan response.ServerSentEvent {
			if id, err := strconv.Atoi(lastEventId); err == nil {
				evtGen.Rewind(id)
			}

			ch := make(chan response.ServerSentEvent)
			go func() {
				ticker := time.NewTicker(100 * time.Millisecond)
				defer ticker.Stop()
				defer close(ch)
				for {
					select {
					case <-ctx.Done():
						return
					case <-ticker.C:
						if evt, ok := evtGen.Next(); ok {
							ch <- evt
						}
					}
				}
			}()

			return ch
		}), nil
	})

	httpServer := http.Server{
		Addr:              ":8080",
		Handler:           gowMux,
		ReadTimeout:       2 * time.Second,
		ReadHeaderTimeout: 1 * time.Second,
		WriteTimeout:      5 * time.Minute, //this long timeout it's necessary for SSE
		IdleTimeout:       30 * time.Second,
		ConnState: func(conn net.Conn, state http.ConnState) {
			if state == http.StateNew {
				log.Printf("New HTTP2 connection: %v", conn.RemoteAddr())
			}
		},
	}
	log.Printf("The server is up and running at address: https://localhost:8080")
	if err := httpServer.ListenAndServeTLS("./examples/certs/key.crt", "./examples/certs/key.key"); err != nil {
		log.Fatalf("Failed to start the server, err: %v", err)
	}

}
