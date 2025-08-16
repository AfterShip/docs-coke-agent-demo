package main

import (
	"context"
	"github.com/cloudwego/eino/schema"
	"io"
	"log"
)

// claudeStreamToolChecker checks if there are tool calls in the stream of messages.
func claudeStreamToolChecker(ctx context.Context, sr *schema.StreamReader[*schema.Message]) (bool, error) {
	defer sr.Close()

	// 获取 responseStreamWriter
	var writer *schema.StreamWriter[*schema.Message]
	if writerVal := ctx.Value("responseStreamWriter"); writerVal != nil {
		if w, ok := writerVal.(*schema.StreamWriter[*schema.Message]); ok {
			writer = w
			log.Printf("Found responseStreamWriter in context")
		} else {
			log.Printf("Warning: responseStreamWriter found in context but is not a valid writer type: %T", writerVal)
		}
	} else {
		log.Printf("No responseStreamWriter found in context")
	}

	for {
		msg, err := sr.Recv()
		if err == io.EOF {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		log.Printf("checker received msg: %s\n", msg.String())
		if msg.ResponseMeta != nil {
			log.Printf("finish: %s", msg.ResponseMeta.FinishReason)
		}

		// 将消息写入 writer（如果存在）
		if writer != nil {
			if closed := writer.Send(msg, nil); closed {
				log.Printf("Writer stream closed unexpectedly")
				break
			}
		}

		if len(msg.ToolCalls) > 0 {
			return true, nil
		}

		if len(msg.Content) == 0 { // skip empty chunks at the front
			continue
		}
	}
	return false, nil

}
