# Quick Start

1- 启动本地 langfuse 服务
```shell
cd scripts/langfuse
docker compose up -d
```

2- 启动 agent 服务
```
export LANGFUSE_PUBLIC_KEY="xxx"
export LANGFUSE_SECRET_KEY="xxxx"
export AM_API_KEY="xxx"
make start
```