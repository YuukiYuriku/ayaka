package toronto

import (
	"net/http"

	"github.com/Unleash/unleash-client-go/v3"
	"github.com/Unleash/unleash-client-go/v3/context"
	"gitlab.com/ayaka/config"
)

type Toggle struct {
	*unleash.Client
	Conf *config.Config `inject:"config"`
}

func (t *Toggle) Startup() error {
	client, err := unleash.NewClient(
		unleash.WithListener(&unleash.DebugListener{}),
		unleash.WithAppName(t.Conf.Toggle.AppName),
		unleash.WithUrl(t.Conf.Toggle.URL),
		unleash.WithCustomHeaders(http.Header{"Authorization": {t.Conf.Toggle.Token}}),
	)

	if err != nil {
		return err
	}

	t.Client = client

	return nil
}

func (f *Toggle) Shutdown() error { return f.Client.Close() }

func (f *Toggle) NewContext(userId string) unleash.FeatureOption {
	ctx := context.Context{
		UserId: userId,
	}

	return unleash.WithContext(ctx)
}
