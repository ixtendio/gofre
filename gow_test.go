package gow

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
		TemplateConfig *TemplateConfig
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
				if c.TemplateConfig != nil {
					t.Errorf("TemplateConfig should be nil")
				}
			},
		},
		{
			name: "test when TemplateConfig values are empty",
			fields: fields{
				ContextPath:    "/a_path",
				TemplateConfig: &TemplateConfig{},
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
				if c.TemplateConfig.TemplatesPathPattern != "resources/templates/*.html" {
					t.Errorf("TemplateConfig.TemplatesPathPattern = %s, want = resources/templates/*.html", c.TemplateConfig.TemplatesPathPattern)
				}
				if c.TemplateConfig.AssetsDirPath != "./resources/assets" {
					t.Errorf("TemplateConfig.AssetsDirPath = %s, want = ./resources/assets", c.TemplateConfig.AssetsDirPath)
				}
				if c.TemplateConfig.AssetsPath != "assets" {
					t.Errorf("TemplateConfig.AssetsPath = %s, want = assets", c.TemplateConfig.AssetsPath)
				}
				if c.TemplateConfig.Template == nil {
					t.Errorf("TemplateConfig.Template is nil, want not nil")
				}
			},
		},
		{
			name: "test when all values are provided",
			fields: fields{
				ContextPath: "/a_path",
				TemplateConfig: &TemplateConfig{
					TemplatesPathPattern: "pattern",
					AssetsDirPath:        "dir_path",
					AssetsPath:           "assets_path",
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
				if c.TemplateConfig.TemplatesPathPattern != "pattern" {
					t.Errorf("TemplateConfig.TemplatesPathPattern = %s, want = pattern", c.TemplateConfig.TemplatesPathPattern)
				}
				if c.TemplateConfig.AssetsDirPath != "dir_path" {
					t.Errorf("TemplateConfig.AssetsDirPath = %s, want = dir_path", c.TemplateConfig.AssetsDirPath)
				}
				if c.TemplateConfig.AssetsPath != "assets_path" {
					t.Errorf("TemplateConfig.AssetsPath = %s, want = assets_path", c.TemplateConfig.AssetsPath)
				}
				if c.TemplateConfig.Template != tmpl {
					t.Errorf("TemplateConfig.Template = %v, want = %v", c.TemplateConfig.Template, tmpl)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				ContextPath:    tt.fields.ContextPath,
				TemplateConfig: tt.fields.TemplateConfig,
				ErrLogFunc:     tt.fields.ErrLogFunc,
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
		want    *Gow
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewGow(tt.args.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewGow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewGow() got = %v, want %v", got, tt.want)
			}
		})
	}
}
