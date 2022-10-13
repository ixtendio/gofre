package gofre

import (
	"html/template"
	"reflect"
	"testing"
)

func TestConfig_setDefaults(t *testing.T) {
	tmpl := template.New("")
	errLogFunc := func(err error) {}
	defaultTemplateFunc = func(templatesPathPattern string) (*template.Template, error) { return template.New(""), nil }
	type fields struct {
		ContextPath    string
		TemplateConfig *ResourcesConfig
		ErrLogFunc     func(err error)
	}
	tests := []struct {
		name       string
		fields     fields
		wantErr    bool
		assertFunc func(t *testing.T, c *Config)
	}{
		{
			name:    "test when all values are empty",
			fields:  fields{},
			wantErr: false,
			assertFunc: func(t *testing.T, c *Config) {
				if c.ContextPath != "/" {
					t.Errorf("ContextPath = %v, want /", c.ContextPath)
				}
				if c.ErrLogFunc == nil {
					t.Errorf("ErrLogFunc should not be nil")
				}
				if c.ResourcesConfig != nil {
					t.Errorf("ResourcesConfig should be nil")
				}
			},
		},
		{
			name: "test when ResourcesConfig values are empty",
			fields: fields{
				ContextPath:    "/a_path",
				TemplateConfig: &ResourcesConfig{},
				ErrLogFunc:     errLogFunc,
			},
			wantErr: false,
			assertFunc: func(t *testing.T, c *Config) {
				if c.ContextPath != "/a_path" {
					t.Errorf("ContextPath = \"%s\", want \"\"", c.ContextPath)
				}
				if reflect.ValueOf(c.ErrLogFunc).Pointer() != reflect.ValueOf(errLogFunc).Pointer() {
					t.Errorf("ErrLogFunc should be errLogFunc")
				}
				if c.ResourcesConfig.TemplatesPathPattern != "resources/templates/*.html" {
					t.Errorf("ResourcesConfig.TemplatesPathPattern = %s, want = resources/templates/*.html", c.ResourcesConfig.TemplatesPathPattern)
				}
				if c.ResourcesConfig.AssetsDirPath != "./resources/assets" {
					t.Errorf("ResourcesConfig.AssetsDirPath = %s, want = ./resources/assets", c.ResourcesConfig.AssetsDirPath)
				}
				if c.ResourcesConfig.AssetsMappingPath != "assets" {
					t.Errorf("ResourcesConfig.AssetsMappingPath = %s, want = assets", c.ResourcesConfig.AssetsMappingPath)
				}
				if c.ResourcesConfig.Template == nil {
					t.Errorf("ResourcesConfig.Template is nil, want not nil")
				}
			},
		},
		{
			name: "test when all values are provided",
			fields: fields{
				ContextPath: "/a_path",
				TemplateConfig: &ResourcesConfig{
					TemplatesPathPattern: "pattern",
					AssetsDirPath:        "dir_path",
					AssetsMappingPath:    "assets_path",
					Template:             tmpl,
				},
				ErrLogFunc: errLogFunc,
			},
			wantErr: false,
			assertFunc: func(t *testing.T, c *Config) {
				if c.ContextPath != "/a_path" {
					t.Errorf("ContextPath = \"%s\", want \"\"", c.ContextPath)
				}
				if reflect.ValueOf(c.ErrLogFunc).Pointer() != reflect.ValueOf(errLogFunc).Pointer() {
					t.Errorf("ErrLogFunc should be errLogFunc")
				}
				if c.ResourcesConfig.TemplatesPathPattern != "pattern" {
					t.Errorf("ResourcesConfig.TemplatesPathPattern = %s, want = pattern", c.ResourcesConfig.TemplatesPathPattern)
				}
				if c.ResourcesConfig.AssetsDirPath != "dir_path" {
					t.Errorf("ResourcesConfig.AssetsDirPath = %s, want = dir_path", c.ResourcesConfig.AssetsDirPath)
				}
				if c.ResourcesConfig.AssetsMappingPath != "assets_path" {
					t.Errorf("ResourcesConfig.AssetsMappingPath = %s, want = assets_path", c.ResourcesConfig.AssetsMappingPath)
				}
				if c.ResourcesConfig.Template != tmpl {
					t.Errorf("ResourcesConfig.Template = %v, want = %v", c.ResourcesConfig.Template, tmpl)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				ContextPath:     tt.fields.ContextPath,
				ResourcesConfig: tt.fields.TemplateConfig,
				ErrLogFunc:      tt.fields.ErrLogFunc,
			}
			if err := c.setDefaults(); (err != nil) != tt.wantErr {
				t.Errorf("setDefaults() error = %v, wantErr %v", err, tt.wantErr)
			}
			tt.assertFunc(t, c)
		})
	}
}

func TestNewSunshine(t *testing.T) {
	type args struct {
		config *Config
	}
	tests := []struct {
		name    string
		args    args
		want    *MuxHandler
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewMuxHandler(tt.args.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewMuxHandler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMuxHandler() got = %v, want %v", got, tt.want)
			}
		})
	}
}
