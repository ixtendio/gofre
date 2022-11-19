package response

import (
	"compress/flate"
	"compress/gzip"
	"errors"
	"github.com/ixtendio/gofre/router/path"

	"net/http"
	"strings"
)

const acceptEncodingHeaderName = "Accept-Encoding"
const contentEncodingHeaderName = "Content-Encoding"

type HttpCompressResponse struct {
	compressionLevel int
	httpResponse     HttpResponse
}

// NewHttpCompressResponse creates a new HttpResponse that provide gzip compression for an HTTP responses as long as the 'Accept-Encoding' request header specify it
// The compressionLevel should be: gzip.DefaultCompression, gzip.NoCompression or any integer value between gzip.BestSpeed and gzip.BestCompression inclusive
func NewHttpCompressResponse(httpResponse HttpResponse, compressionLevel int) (*HttpCompressResponse, error) {
	if compressionLevel < gzip.DefaultCompression || compressionLevel > gzip.BestCompression {
		return nil, errors.New("compression level not supported")
	}
	return &HttpCompressResponse{
		compressionLevel: compressionLevel,
		httpResponse:     httpResponse,
	}, nil
}

func (r *HttpCompressResponse) StatusCode() int {
	return r.httpResponse.StatusCode()
}

func (r *HttpCompressResponse) Headers() HttpHeaders {
	return r.httpResponse.Headers()
}

func (r *HttpCompressResponse) Cookies() HttpCookies {
	return r.httpResponse.Cookies()
}

func (r *HttpCompressResponse) Write(w http.ResponseWriter, req path.MatchingContext) error {
	// detect what compression algorithm to use
	compressAlg := getCompressionAlgorithmFromHeaderValue(req.R.Header.Get(acceptEncodingHeaderName))

	if compressAlg == "gzip" {
		cw, err := gzip.NewWriterLevel(w, r.compressionLevel)
		if err == nil {
			delete(r.Headers(), acceptEncodingHeaderName)
			w.Header().Set(contentEncodingHeaderName, compressAlg)
			w = &compressResponseWriter{
				origResponseWriter: w,
				compressWriter:     cw,
			}
			defer cw.Close()
		}
	} else if compressAlg == "deflate" {
		cw, err := flate.NewWriter(w, r.compressionLevel)
		if err == nil {
			delete(r.Headers(), acceptEncodingHeaderName)
			w.Header().Set(contentEncodingHeaderName, compressAlg)
			w = &compressResponseWriter{
				origResponseWriter: w,
				compressWriter:     cw,
			}
			defer cw.Close()
		}
	}
	return r.httpResponse.Write(w, req)
}

func getCompressionAlgorithmFromHeaderValue(acceptEncodingHeaderNameVal string) string {
	encArr := strings.Split(strings.ToLower(acceptEncodingHeaderNameVal), ",")
	for _, enc := range encArr {
		enc = strings.TrimSpace(enc)
		if enc == "gzip" || enc == "compress" {
			return "gzip"
		}
	}
	for _, enc := range encArr {
		enc = strings.TrimSpace(enc)
		if enc == "deflate" {
			return "deflate"
		}
	}
	return ""
}

type compressWriter interface {
	Write(bytes []byte) (int, error)
	Flush() error
}

type compressResponseWriter struct {
	origResponseWriter http.ResponseWriter
	compressWriter     compressWriter
}

func (c *compressResponseWriter) Header() http.Header {
	return c.origResponseWriter.Header()
}

func (c *compressResponseWriter) Write(bytes []byte) (int, error) {
	return c.compressWriter.Write(bytes)
}

func (c *compressResponseWriter) WriteHeader(statusCode int) {
	c.origResponseWriter.WriteHeader(statusCode)
}

func (c *compressResponseWriter) Flush() {
	c.compressWriter.Flush()
}
