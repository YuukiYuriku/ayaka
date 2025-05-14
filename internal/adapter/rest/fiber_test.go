package rest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/ayaka/config"
)

func TestFiber_Startup(t *testing.T) {
	// Create a new instance of the Fiber struct
	f := &Fiber{
		Conf: &config.Config{
			Http: config.HttpConfig{
				ReadTimeout:  10,
				WriteTimeout: 10,
			},
		},
	}

	// Call the Startup method
	err := f.Startup()

	// Assert that there is no error
	assert.NoError(t, err)
}

func TestFiber_Shutdown(t *testing.T) {
	// Create a new instance of the Fiber struct
	f := &Fiber{
		Conf: &config.Config{
			Http: config.HttpConfig{
				ReadTimeout:  10,
				WriteTimeout: 10,
			},
		},
	}

	// Call the Startup method
	err := f.Startup()
	assert.NoError(t, err)

	// Call the Shutdown method
	err = f.Shutdown()
	// Assert that there is no error
	assert.NoError(t, err)
}
