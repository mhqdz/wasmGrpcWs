声明:
	本项目拷贝自 tarndt的[wasmws ](https://github.com/tarndt/wasmws)仅针对项目使用中遇到跨域请求问题(`err: not authorized for Host`)稍微做了一些修改 (我不确定该错误是否是由我的配置错误引起的)

​	我的做法是在函数`NewWebSocketListener()`中传入了`websocket.AcceptOptions` 并在`websocket.Accept`时使用该opts



示例demo的运行方法:

在`demo/server`目录下使用:

```sh
make run
```

---

Declaration:
    This project is a copy of tarndt's [wasmws](https://github.com/tarndt/wasmws) for the purpose of solving the cross-domain request problem (err: not authorized for Host) with some modifications (I'm not sure if the error is caused by my configuration error).

My approach is to pass `websocket.AcceptOptions` to the `function NewWebSocketListener()` and use it in `websocket.Accept().`

Example demo running method:

In the demo/server directory, use:

```sh
make run
```
