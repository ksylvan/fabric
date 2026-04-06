package restapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

const maxJSONRequestBytes = 16 << 20

var errMultipleJSONValues = errors.New("request body must contain a single JSON object")

func decodeStrictJSON(c *gin.Context, dest any) error {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxJSONRequestBytes)

	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	decoder.UseNumber()

	if err := decoder.Decode(dest); err != nil {
		return err
	}

	var extra json.RawMessage
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return errMultipleJSONValues
		}
		return err
	}

	return nil
}
