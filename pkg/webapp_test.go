package gowebapp

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewWebApp(t *testing.T) {
	var tests = []struct {
		name string
		environment string
		expectedScope Scope

	}{
		{
			name: "NewWebApp Ok",
			environment: "dummy-env",
			expectedScope: Scope{Environment: "dummy-env"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			webapp := NewWebApp(tt.environment)
			require.Equal(t, tt.expectedScope, webapp.Scope)
			require.NotEmpty(t, webapp.Router)
			require.NotEmpty(t, webapp)
		})
	}
}

func TestWebApp_Run(t *testing.T) {
	environment := "dummy-env"
	webapp1 := NewWebApp(environment)

	go func() {
		err := webapp1.Run()
		require.NoError(t, err)
	}()

	go func() {
		webapp2 := NewWebApp(environment)
		err := webapp2.Run()
		require.Errorf(t, err, "port already used by another WebApp")
	}()
}

