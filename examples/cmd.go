package main

import (
	"context"
	"github.com/ixtendio/gofre"
	"github.com/ixtendio/gofre/auth"
	"github.com/ixtendio/gofre/auth/oauth"
	"github.com/ixtendio/gofre/cache"
	"github.com/ixtendio/gofre/handler"
	"github.com/ixtendio/gofre/middleware"
	"github.com/ixtendio/gofre/request"
	"github.com/ixtendio/gofre/response"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {

	ctx := context.Background()
	gofreConfig := &gofre.Config{
		CaseInsensitivePathMatch: false,
		ContextPath:              "",
		ResourcesConfig: &gofre.ResourcesConfig{
			TemplatesPathPattern: "resources/templates/*.html",
			AssetsDirPath:        "./resources/assets",
		},
		ErrLogFunc: func(err error) {
			log.Printf("An error occurred in the GOFre framework: %v", err)
		},
	}
	gofreMux, err := gofre.NewMuxHandler(gofreConfig)
	if err != nil {
		log.Fatalf("Failed to create GOFre mux handler, err: %v", err)
	}
	gofreMux.CommonPreMiddlewares(middleware.Panic(), middleware.ErrJsonResponse())

	// template example
	gofreMux.HandleGet("/", func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
		templateData := struct{}{}
		return response.TemplateHttpResponseOK(gofreConfig.ResourcesConfig.Template, "index.html", templateData), nil
	})

	// OAUTH2 flow with user details extraction
	gofreMux.HandleOAUTH2(oauth.Config{
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
	gofreMux.HandleGet("/text/{plain}", func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
		return response.PlainTextHttpResponseOK("Text plain response"), nil
	})
	gofreMux.HandleGet("/text/*", func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
		return response.PlainTextHttpResponseOK("Text plain response"), nil
	})

	// TEXT plain response
	gofreMux.HandleGet("/text/plain", func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
		return response.PlainTextHttpResponseOK("Text plain response"), nil
	})

	// HTML response
	gofreMux.HandleGet("/text/html", func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
		return response.HtmlHttpResponseOK("<!DOCTYPE html><html><body><h1>HTML example</h1></body></html>"), nil
	})

	// JSON with vars path
	gofreMux.HandleGet("/json/{user}/{id}", func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
		return response.JsonHttpResponseOK(r.UriVars), nil
	})

	// document download example
	gofreMux.HandleGet("/download", func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
		f, err := os.Open("./resources/assets/image.png")
		if err != nil {
			return nil, err
		}
		return response.StreamHttpResponse(f, "image/png"), nil
	})

	// template example
	gofreMux.HandleGet("/tmpl/{tmplName}", func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
		templateName := r.UriVars["tmplName"] + ".html"
		return response.TemplateHttpResponseOK(gofreConfig.ResourcesConfig.Template, templateName, nil), nil
	})

	// SSE example
	evtGen := NewEventGenerator(ctx, 10)
	gofreMux.HandleGet("/sse", func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
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

	// Authorization example
	gofreMux.HandleGet("/security/authorize/{permission}", func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
		return response.JsonHttpResponseOK(map[string]string{"authorized": "true"}), nil
	}, func(handler handler.Handler) handler.Handler {
		// authentication provider
		return func(ctx context.Context, req *request.HttpRequest) (resp response.HttpResponse, err error) {
			permission, err := auth.ParsePermission("domain/subdomain/resource:" + req.UriVars["permission"])
			if err != nil {
				return nil, err
			}
			ctx = context.WithValue(ctx, auth.SecurityPrincipalCtxKey, auth.User{
				Groups: []auth.Group{{
					Roles: []auth.Role{{
						AllowedPermissions: []auth.Permission{permission},
					}},
				}},
			})
			return handler(ctx, req)
		}
	}, middleware.AuthorizeAll(auth.Permission{Scope: "domain/subdomain/resource", Access: auth.AccessDelete}))

	httpServer := http.Server{
		Addr:              ":8080",
		Handler:           gofreMux,
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
	if err := httpServer.ListenAndServeTLS("./certs/key.crt", "./certs/key.key"); err != nil {
		log.Fatalf("Failed to start the server, err: %v", err)
	}

}
