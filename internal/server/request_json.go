package restapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

const maxRequestBodyBytes = 16 << 20

var errMultipleJSONValues = errors.New("request body must contain a single JSON object")
var errRequestBodyTooLarge = errors.New("request body too large")

func decodeStrictJSON(c *gin.Context, dest any) error {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRequestBodyBytes)

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

func readLimitedRequestBody(c *gin.Context) ([]byte, error) {
	limitedBody := http.MaxBytesReader(c.Writer, c.Request.Body, maxRequestBodyBytes)
	defer limitedBody.Close()

	c.Request.Body = limitedBody
	return io.ReadAll(limitedBody)
}

func isRequestBodyTooLarge(err error) bool {
	var maxBytesErr *http.MaxBytesError
	return errors.As(err, &maxBytesErr)
}
