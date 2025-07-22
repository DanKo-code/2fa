package external

import (
	"log/slog"
	"net/http"
)

func SafeCloseBody(resp *http.Response) {
	if resp != nil && resp.Body != nil {
		if err := resp.Body.Close(); err != nil {
			slog.Error("failed to close response body", "err", err)
		}
	}
}
