package main

import (
	"context"
	"github.com/ixtendio/gofre"
	"github.com/ixtendio/gofre/auth"
	"github.com/ixtendio/gofre/auth/oauth"
	"github.com/ixtendio/gofre/cache"
	"github.com/ixtendio/gofre/middleware"
	"github.com/ixtendio/gofre/router/path"

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
	gofreMux, err := gofre.NewMuxHandlerWithDefaultConfigAndTemplateSupport()
	if err != nil {
		log.Fatalf("Failed to create GOFre mux handler, err: %v", err)
	}
	gofreMux.CommonMiddlewares(middleware.PanicRecover(), middleware.ErrJsonResponse())

	// template example
	gofreMux.HandleGet("/", func(ctx context.Context, mc path.MatchingContext) (response.HttpResponse, error) {
		templateData := struct{}{}
		return response.TemplateHttpResponseOK(gofreMux.ExecutableTemplate(), "index.html", templateData), nil
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
	}, func(ctx context.Context, mc path.MatchingContext) (response.HttpResponse, error) {
		accessToken := oauth.GetAccessTokenFromContext(ctx)
		securityPrincipal := auth.GetSecurityPrincipalFromContext(ctx)
		return response.JsonHttpResponseOK(map[string]interface{}{
			"accessToken":       accessToken,
			"authenticatedUser": securityPrincipal,
		}), nil
	}, nil, nil)

	// TEXT plain response
	textRouter := gofreMux.RouteUsingPathPrefix("/text")
	textRouter.HandleGet("/{plain}", func(ctx context.Context, mc path.MatchingContext) (response.HttpResponse, error) {
		return response.PlainTextHttpResponseOK("Text plain response"), nil
	})
	textRouter.HandleGet("/*", func(ctx context.Context, mc path.MatchingContext) (response.HttpResponse, error) {
		return response.PlainTextHttpResponseOK("Text plain response"), nil
	})
	// TEXT plain response
	textRouter.HandleGet("/plain", func(ctx context.Context, mc path.MatchingContext) (response.HttpResponse, error) {
		return response.PlainTextHttpResponseOK("Text plain response"), nil
	})
	// HTML response
	textRouter.HandleGet("/html", func(ctx context.Context, mc path.MatchingContext) (response.HttpResponse, error) {
		return response.HtmlHttpResponseOK("<!DOCTYPE html><html><body><h1>HTML example</h1></body></html>"), nil
	})

	// JSON with vars path
	gofreMux.HandleGet("/json/{user}/{id}", func(ctx context.Context, mc path.MatchingContext) (response.HttpResponse, error) {
		return response.JsonHttpResponseOK(map[string]string{
			"user": mc.PathVar("user"),
			"id":   mc.PathVar("id"),
		}), nil
	})

	// document download example
	gofreMux.HandleGet("/download", func(ctx context.Context, mc path.MatchingContext) (response.HttpResponse, error) {
		f, err := os.Open("./resources/assets/image.png")
		if err != nil {
			return nil, err
		}
		return response.StreamHttpResponse("image/png", f), nil
	})

	// template example
	gofreMux.HandleGet("/tmpl/{tmplName}", func(ctx context.Context, mc path.MatchingContext) (response.HttpResponse, error) {
		templateName := mc.PathVar("tmplName") + ".html"
		return response.TemplateHttpResponseOK(gofreMux.ExecutableTemplate(), templateName, nil), nil
	})

	// SSE example
	evtGen := NewEventGenerator(ctx, 10)
	gofreMux.HandleGet("/sse", func(ctx context.Context, mc path.MatchingContext) (response.HttpResponse, error) {
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
	gofreMux.HandleGet("/security/authorize/{permission}", func(ctx context.Context, mc path.MatchingContext) (response.HttpResponse, error) {
		return response.JsonHttpResponseOK(map[string]string{"authorized": "true"}), nil
	}, middleware.SecurityPrincipalSupplier(func(ctx context.Context, mc path.MatchingContext) (auth.SecurityPrincipal, error) {
		permission, err := auth.ParsePermission("domain/subdomain/resource:" + mc.PathVar("permission"))
		if err != nil {
			return nil, err
		}
		return auth.User{
			Groups: []auth.Group{{
				Roles: []auth.Role{{
					AllowedPermissions: []auth.Permission{permission},
				}},
			}},
		}, nil
	}), middleware.AuthorizeAll(auth.Permission{Scope: "domain/subdomain/resource", Access: auth.AccessDelete}))

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
