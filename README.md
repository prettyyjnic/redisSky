# redisSky
redis web 管理工具, 参考 redisMaster, 该项目主要是因为使用 redisMaster 处理大数据量的时候容易卡死, 很不爽, 所以用 ivew + golang 仿写了一个 web 界面的 redis 管理工具

# 不兼容 ie, 作为一个程序员不应该用ie

```
前端使用 iview + socket.io
后端使用 golang + socket.io

Demo 地址：
http://59.110.239.205

联系邮箱：
prettyyjnic@qq.com

使用：
# 
cd $项目根目录/frontend 
npm install # 安装依赖
npm run build # 前端代码编译
cd $项目根目录/backend/bin && go run start.go # 启动服务


```