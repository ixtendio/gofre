package response

import (
	"compress/flate"
	"compress/gzip"
	"github.com/ixtendio/gofre/request"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

const bigText = `Lorem ipsum dolor sit amet, consectetur adipiscing elit. Morbi fermentum massa vitae metus fringilla efficitur. Vestibulum viverra fringilla mollis. Nulla euismod tellus ac turpis convallis pulvinar. In aliquet posuere libero congue venenatis. Vestibulum laoreet nec justo nec luctus. Fusce dictum justo vitae auctor rutrum. Aenean a commodo risus, sit amet faucibus urna. Aenean et sem quis lacus efficitur gravida nec vitae lorem. Ut accumsan suscipit erat ac porttitor. Donec pharetra lorem a turpis egestas placerat. Sed quis magna non nulla ultricies auctor eget eu eros. Cras suscipit semper orci ut convallis.
Suspendisse viverra sollicitudin mattis. In finibus non ex et auctor. Aenean mattis neque urna, eget ullamcorper erat sollicitudin sed. Mauris scelerisque leo diam, ut interdum arcu lobortis vulputate. Sed sed mi justo. Aliquam rutrum ipsum vel congue dignissim. In ipsum sapien, molestie in dictum id, suscipit quis risus. Aliquam ac iaculis nibh. Cras ante nibh, hendrerit vel tortor et, scelerisque lobortis elit. Proin volutpat neque eget interdum rutrum. Curabitur aliquam turpis eu scelerisque porta. Aenean egestas lacus quis porta aliquam. Donec imperdiet purus vitae leo scelerisque imperdiet. Nulla nunc ante, cursus eu interdum eu, imperdiet sit amet diam. Morbi egestas metus non dui semper pulvinar. Morbi a nisl nec erat volutpat sagittis nec sit amet quam.
Nullam a augue non libero viverra efficitur in et mi. Nam id ex id elit lacinia vestibulum eu bibendum justo. Nullam nibh ante, malesuada eu velit nec, maximus imperdiet nisi. Nulla libero nunc, sodales nec vulputate quis, feugiat nec tortor. Morbi tristique sapien a velit vulputate pulvinar. Pellentesque placerat odio ut pretium dapibus. Integer varius finibus turpis, eget tempus ex ultricies sed. Fusce interdum feugiat velit varius efficitur. Nam ac nulla molestie, congue metus at, euismod ipsum.
Morbi convallis quis augue sollicitudin dapibus. Sed feugiat felis a ex lacinia ornare. Proin venenatis lectus eu turpis fermentum gravida. Aenean ultrices erat velit, sit amet posuere tortor tempor sit amet. Vivamus nulla nunc, egestas et elementum eu, rhoncus nec dui. Aliquam molestie augue eu enim dapibus aliquam. Nam tincidunt massa eu nibh pharetra, commodo facilisis enim tincidunt. Fusce sit amet convallis est, in imperdiet elit. Fusce elit odio, hendrerit et urna tempus, iaculis malesuada erat. In hac habitasse platea dictumst. Aliquam et enim a quam condimentum sollicitudin id a nunc. Nulla maximus tortor id purus gravida feugiat. Duis consectetur auctor dui et molestie. Curabitur id tincidunt lectus. Curabitur et dictum eros. Etiam consequat, quam sit amet sodales viverra, erat purus hendrerit erat, ut dapibus lorem orci ut ligula.
Phasellus viverra et tortor eu laoreet. Vestibulum tempor sed nisi vitae iaculis. In in tortor sit amet velit laoreet sagittis in quis ligula. Nam a quam dui. Nunc odio turpis, blandit vitae justo non, efficitur commodo massa. Maecenas efficitur auctor est, vel bibendum lectus. Suspendisse ut odio quis est blandit hendrerit sed sed massa. Ut id lectus ac est sollicitudin feugiat ut non magna.
Curabitur ultricies ex eleifend porta ultricies. Curabitur convallis nec massa id auctor. Aliquam vehicula vestibulum mauris, at hendrerit enim auctor ut. Suspendisse potenti. Aliquam vel lobortis purus. Cras massa sapien, sollicitudin vel scelerisque scelerisque, tempus a mi. Vestibulum quis sodales eros. Nulla a sapien tellus. Quisque a nibh vel arcu aliquet lobortis sed a orci. Duis fermentum fringilla sapien, sit amet faucibus ipsum facilisis in.
Vestibulum imperdiet ante sed varius blandit. Etiam vitae quam at lorem blandit fringilla sit amet eget erat. In eu lacus commodo, vehicula risus quis, sollicitudin nisl. Curabitur cursus porta purus mattis pharetra. Fusce fringilla massa turpis, ac euismod purus feugiat et. Sed efficitur sem ut efficitur mattis. Proin malesuada nibh ut enim porta, a tempor lectus pellentesque. Aliquam id nisi eget tellus rutrum vestibulum eu sed eros. Cras sit amet fermentum dui, vel gravida sapien. Aenean pretium tempus ipsum at dictum. Cras velit felis, aliquam in elit viverra, porttitor malesuada purus. In rutrum massa id nisl dictum viverra.
Ut commodo, nisi eu eleifend condimentum, neque dolor congue mi, in blandit arcu nibh quis erat. Ut finibus volutpat leo. Vivamus varius dolor lacinia est rutrum, nec sodales ipsum faucibus. Vivamus egestas arcu diam, sed rutrum dolor gravida quis. Vivamus posuere velit ut urna congue mollis. In est quam, condimentum ac nibh sit amet, luctus aliquam mauris. Nam id ullamcorper velit, in commodo erat. Interdum et malesuada fames ac ante ipsum primis in faucibus. Nullam dapibus eros a purus euismod, ac tincidunt tellus elementum. Nullam facilisis vestibulum augue, eget elementum nibh. Sed at scelerisque libero. Nam lacinia laoreet varius. Duis faucibus ex sed porta gravida.
Cras auctor laoreet lorem id laoreet. Nulla ac euismod felis. Sed pulvinar venenatis nibh. Pellentesque rhoncus erat eget diam luctus, non ultrices eros iaculis. Vivamus sagittis mauris mi. Sed sed diam id augue vehicula tempus sed a ante. Nam dolor magna, dignissim non lobortis ac, dapibus ac nisl. Pellentesque vitae massa consectetur, consequat nisi vel, consequat nibh. Suspendisse lectus nibh, consectetur at nulla ut, condimentum maximus ipsum. Vivamus congue dui sit amet augue malesuada accumsan. Phasellus semper dignissim ligula vel auctor. Quisque lobortis elit id nulla dapibus, blandit posuere nisl ullamcorper. Ut quis est lobortis, interdum diam in, vulputate felis. Sed nulla ipsum, convallis in libero in, cursus aliquam nisl. Phasellus quis auctor orci. Donec ac eleifend tortor.
Vivamus ac ultricies magna, ac porttitor metus. Cras rutrum augue eget pretium porttitor. Proin euismod velit non lacus gravida ultricies. Proin quis maximus tortor. Praesent tincidunt dui quis leo posuere, nec egestas dui suscipit. Proin pellentesque, orci a elementum ultrices, orci lectus efficitur turpis, vitae luctus mauris dui a nibh. Sed nulla sem, elementum a cursus sed, eleifend at ligula. Aliquam pretium lorem sed dapibus ullamcorper. Cras tempus nisi sed eros malesuada, id finibus quam vulputate. Vivamus id aliquam lectus. Nam hendrerit elit feugiat felis accumsan hendrerit. Duis ultrices leo eu dui ultricies tincidunt. Sed felis felis, ullamcorper eu neque non, accumsan imperdiet dui. Ut accumsan eu leo sit amet porta. Nunc bibendum arcu sit amet suscipit porta. Aenean egestas mi non volutpat mattis. Cras dapibus hendrerit varius. Sed feugiat scelerisque sodales. Quisque sodales cursus facilisis. Duis.`

func TestHttpCompressResponse_Write(t *testing.T) {
	type args struct {
		w   *httptest.ResponseRecorder
		req *http.Request
	}
	tests := []struct {
		name                string
		args                args
		wantCompressionType string
	}{
		{
			name: "gzip compression",
			args: args{
				w:   httptest.NewRecorder(),
				req: &http.Request{Header: http.Header{acceptEncodingHeaderName: {"gzip"}}},
			},
			wantCompressionType: "gzip",
		},
		{
			name: "gzip compression when header is compress",
			args: args{
				w:   httptest.NewRecorder(),
				req: &http.Request{Header: http.Header{acceptEncodingHeaderName: {"compress"}}},
			},
			wantCompressionType: "gzip",
		},
		{
			name: "flate compression",
			args: args{
				w:   httptest.NewRecorder(),
				req: &http.Request{Header: http.Header{acceptEncodingHeaderName: {"deflate"}}},
			},
			wantCompressionType: "deflate",
		},
		{
			name: "no compression",
			args: args{
				w:   httptest.NewRecorder(),
				req: &http.Request{Header: http.Header{}},
			},
			wantCompressionType: "",
		},
		{
			name: "no compression when header value is wrong",
			args: args{
				w:   httptest.NewRecorder(),
				req: &http.Request{Header: http.Header{acceptEncodingHeaderName: {"lz4"}}},
			},
			wantCompressionType: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &HttpCompressResponse{
				compressionLevel: gzip.DefaultCompression,
				httpResponse:     PlainTextHttpResponseOK(bigText),
			}
			if err := r.Write(tt.args.w, request.NewHttpRequest(tt.args.req)); err != nil {
				t.Fatalf("Write() returned error: %v", err)
			}
			decompressedValue, err := decompress(tt.wantCompressionType, tt.args.w.Body)
			if err != nil {
				t.Fatalf("Write() decompress error: %v", err)
			}
			if decompressedValue != bigText {
				t.Fatalf("Write() got decompressed: '%s'", decompressedValue)
			}
			contentEncodingHeaderValue := tt.args.w.Header().Get(contentEncodingHeaderName)
			if tt.wantCompressionType != "" {
				if contentEncodingHeaderValue != tt.wantCompressionType {
					t.Fatalf("Write() Content-Encoding header value, got: %s want: %s ", contentEncodingHeaderValue, tt.wantCompressionType)
				}
			} else {
				if contentEncodingHeaderValue != "" {
					t.Fatalf("Write() Content-Encoding header value, got: %s want to be empty ", contentEncodingHeaderValue)
				}
			}
		})
	}
}

func decompress(compressType string, compressedData io.Reader) (string, error) {
	if compressType == "gzip" {
		r, err := gzip.NewReader(compressedData)
		if err != nil {
			return "", err
		}
		output, err := io.ReadAll(r)
		if err != nil {
			return "", err
		}
		return string(output), nil
	} else if compressType == "deflate" {
		output, err := io.ReadAll(flate.NewReader(compressedData))
		if err != nil {
			return "", err
		}
		return string(output), nil
	} else {
		output, err := io.ReadAll(compressedData)
		if err != nil {
			return "", err
		}
		return string(output), nil
	}
}

func TestNewHttpCompressResponse(t *testing.T) {
	defaultHttpResponse := PlainTextHttpResponseOK("ok")
	type args struct {
		compressionLevel int
	}
	tests := []struct {
		name    string
		args    args
		want    *HttpCompressResponse
		wantErr bool
	}{
		{
			name: "no error",
			args: args{
				compressionLevel: 0,
			},
			want: &HttpCompressResponse{
				httpResponse:     defaultHttpResponse,
				compressionLevel: 0,
			},
			wantErr: false,
		},
		{
			name: "error if compression level not supported",
			args: args{
				compressionLevel: -2,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewHttpCompressResponse(defaultHttpResponse, tt.args.compressionLevel)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewHttpCompressResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("NewHttpCompressResponse() got = %v, want %v", got, tt.want)
				}
				if !reflect.DeepEqual(got.Headers(), defaultHttpResponse.Headers()) {
					t.Errorf("NewHttpCompressResponse() got headers = %v, want %v", got.Headers(), defaultHttpResponse.Headers())
				}
				if !reflect.DeepEqual(got.StatusCode(), defaultHttpResponse.StatusCode()) {
					t.Errorf("NewHttpCompressResponse() got statusCode = %v, want %v", got.StatusCode(), defaultHttpResponse.StatusCode())
				}
				if !reflect.DeepEqual(got.Cookies(), defaultHttpResponse.Cookies()) {
					t.Errorf("NewHttpCompressResponse() got cookies = %v, want %v", got.Cookies(), defaultHttpResponse.Cookies())
				}
			}
		})
	}
}
