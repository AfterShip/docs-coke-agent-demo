package aftership_genkit

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/anthropics/anthropic-sdk-go/option"
	"io"
	"net/http"
	"strings"
)

// streamCorrectionMiddleware is a middleware that wraps the response body
// AfterShip AIGC API stream 模式下的 response 行为与 Anthropic 的不一致, 少了 event: 行，
// 为了能直接使用 Anthropic 的流式处理逻辑，需要在 AfterShip 的流式响应中添加缺失的 event: 行
var streamCorrectionMiddleware = func(request *http.Request, next option.MiddlewareNext) (*http.Response, error) {
	resp, err := next(request)
	if err != nil {
		return resp, err
	}

	// Check if this is a streaming request
	if resp.Header.Get("Content-Type") == "text/event-stream" {
		// Wrap the response body to add missing event lines
		resp.Body = &streamBodyWrapper{
			originalBody: resp.Body,
		}
	}

	return resp, err
}

// streamBodyWrapper wraps the response body to add missing event lines
type streamBodyWrapper struct {
	originalBody io.ReadCloser
	buffer       bytes.Buffer
	scanner      *bufio.Scanner
	initialized  bool
}

func (w *streamBodyWrapper) Read(p []byte) (n int, err error) {
	if !w.initialized {
		w.scanner = bufio.NewScanner(w.originalBody)
		w.initialized = true
	}

	// If we have data in buffer, read from it first
	if w.buffer.Len() > 0 {
		return w.buffer.Read(p)
	}

	// Read next line from original body
	if w.scanner.Scan() {
		line := w.scanner.Text()

		// If line starts with "data:" and the previous line wasn't an "event:" line,
		// we need to add the missing event line
		if strings.HasPrefix(line, "data:") {
			// Check what type of event this is based on the data content
			var eventType string
			if strings.Contains(line, `"type":"message_start"`) {
				eventType = "message_start"
			} else if strings.Contains(line, `"type":"content_block_start"`) {
				eventType = "content_block_start"
			} else if strings.Contains(line, `"type":"ping"`) {
				eventType = "ping"
			} else if strings.Contains(line, `"type":"content_block_delta"`) {
				eventType = "content_block_delta"
			} else if strings.Contains(line, `"type":"content_block_stop"`) {
				eventType = "content_block_stop"
			} else if strings.Contains(line, `"type":"message_delta"`) {
				eventType = "message_delta"
			} else if strings.Contains(line, `"type":"message_stop"`) {
				eventType = "message_stop"
			} else if strings.Contains(line, `"type":"error"`) {
				eventType = "error"
			}

			if eventType != "" {
				w.buffer.WriteString(fmt.Sprintf("event: %s\n", eventType))
			}
		}

		// Add the original line
		w.buffer.WriteString(line + "\n")

		return w.buffer.Read(p)
	}

	// Check for scanning errors
	if err := w.scanner.Err(); err != nil {
		return 0, err
	}

	// EOF
	return 0, io.EOF
}

func (w *streamBodyWrapper) Close() error {
	return w.originalBody.Close()
}
