// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package awsprometheusremotewriteexporter provides a Prometheus Remote Write Exporter with AWS Sigv4 authentication
package awsprometheusremotewriteexporter

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configcheck"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/config/configmodels"
	"go.opentelemetry.io/collector/config/configtls"
	"go.uber.org/zap"
)

func TestType(t *testing.T) {
	af := NewFactory()
	assert.Equal(t, af.Type(), configmodels.Type(typeStr))
}

//Tests whether or not the default Exporter factory can instantiate a properly interfaced Exporter with default conditions
func TestCreateDefaultConfig(t *testing.T) {
	af := NewFactory()
	cfg := af.CreateDefaultConfig()
	assert.NotNil(t, cfg, "failed to create default config")
	assert.NoError(t, configcheck.ValidateConfig(cfg))
}

//Tests whether or not a correct Metrics Exporter from the default Config parameters
func TestCreateMetricsExporter(t *testing.T) {
	af := NewFactory()
	validConfigWithAuth := af.CreateDefaultConfig().(*Config)
	validConfigWithAuth.AuthConfig = AuthConfig{Region: "region", Service: "service"}

	// Some form of AWS credentials chain required to test valid auth case
	os.Setenv("AWS_ACCESS_KEY", "string")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "string2")

	invalidConfigWithAuth := af.CreateDefaultConfig().(*Config)
	invalidConfigWithAuth.AuthConfig = AuthConfig{Region: "", Service: "service"}

	invalidConfig := af.CreateDefaultConfig().(*Config)
	invalidConfig.HTTPClientSettings = confighttp.HTTPClientSettings{}

	invalidTLSConfig := af.CreateDefaultConfig().(*Config)
	invalidTLSConfig.HTTPClientSettings.TLSSetting = configtls.TLSClientSetting{
		TLSSetting: configtls.TLSSetting{
			CAFile:   "non-existent file",
			CertFile: "",
			KeyFile:  "",
		},
		Insecure:   false,
		ServerName: "",
	}

	tests := []struct {
		name        string
		cfg         configmodels.Exporter
		params      component.ExporterCreateParams
		returnError bool
	}{
		{"success_case_default",
			af.CreateDefaultConfig(),
			component.ExporterCreateParams{Logger: zap.NewNop()},
			false,
		},
		{"success_case_with_auth",
			validConfigWithAuth,
			component.ExporterCreateParams{Logger: zap.NewNop()},
			false,
		},
		{"invalid_auth_case",
			invalidConfigWithAuth,
			component.ExporterCreateParams{Logger: zap.NewNop()},
			true,
		},
		{"invalid_config_case",
			invalidConfig,
			component.ExporterCreateParams{Logger: zap.NewNop()},
			true,
		},
		{"invalid_tls_config_case",
			invalidTLSConfig,
			component.ExporterCreateParams{Logger: zap.NewNop()},
			true,
		},
	}
	// run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := af.CreateMetricsExporter(context.Background(), tt.params, tt.cfg)
			if tt.returnError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
