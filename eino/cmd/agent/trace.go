package main

import (
	"github.com/cloudwego/eino-ext/callbacks/langfuse"
	"github.com/cloudwego/eino/callbacks"
	"os"
)

func setupTracing() {
	publicKey := os.Getenv("LANGFUSE_PUBLIC_KEY")
	secretKey := os.Getenv("LANGFUSE_SECRET_KEY")

	cbh, _ := langfuse.NewLangfuseHandler(&langfuse.Config{
		Host:      "http://127.0.0.1:3000",
		Name:      "eino-demo",
		Release:   "v1.0.0",
		Public:    true,
		SecretKey: secretKey,
		PublicKey: publicKey,
	})

	callbacks.AppendGlobalHandlers(cbh)
}
