package ollama

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/danielmiessler/fabric/internal/i18n"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadImageBytes_DataURLValidationErrorsAreLocalized(t *testing.T) {
	_, err := i18n.Init("en")
	require.NoError(t, err)

	client := &Client{}

	_, err = client.loadImageBytes(context.Background(), "data:image/png;base64")
	require.Error(t, err)
	assert.Equal(t, i18n.T("ollama_invalid_data_url_format"), err.Error())

	_, err = client.loadImageBytes(context.Background(), "data:image/png;base64,%%%%")
	require.Error(t, err)
	assert.True(t, strings.HasPrefix(err.Error(), strings.Split(i18n.T("ollama_failed_decode_data_url"), "%v")[0]))
}

func TestLoadImageBytes_HTTPFetchErrorIsLocalized(t *testing.T) {
	_, err := i18n.Init("en")
	require.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	client := &Client{httpClient: server.Client()}

	_, err = client.loadImageBytes(context.Background(), server.URL+"/image.png")
	require.Error(t, err)
	assert.Equal(t,
		fmt.Sprintf(i18n.T("ollama_failed_fetch_image"), server.URL+"/image.png", "500 Internal Server Error"),
		err.Error(),
	)
}

func TestLoadImageBytes_DataURLSuccess(t *testing.T) {
	_, err := i18n.Init("en")
	require.NoError(t, err)

	client := &Client{}
	expected := []byte("hello world")
	dataURL := "data:image/png;base64," + base64.StdEncoding.EncodeToString(expected)

	got, err := client.loadImageBytes(context.Background(), dataURL)
	require.NoError(t, err)
	assert.Equal(t, expected, got)
}
