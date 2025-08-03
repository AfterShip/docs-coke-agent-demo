# 启动项目的方式
0- 安装依赖
```shell
npm install -g genkit-cli
```
这个仅影响后面的 debug 命令启动本地调试工具，实际代码开发不依赖这个工具。

1- 配置环境变量
```shell
export AM_API_KEY=xxxx
```


2- 启动项目
默认启动的是测试环境，so, 请配置测试环境的 API KEY
```shell
make debug
```
这个命令会在本地启动一个调试服务器，监听在 `localhost:4000`，直接网页打开：http://localhost:4000/


3- 编译项目
```shell
make build
```
