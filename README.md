## 测试环境
机器配置：4C8G

能跑 2.5 个机器人，超过此上限则会出现 Client 内存泄露（消息处理不完，产生积压）。

## 多机部署环境
连接数共计约 10W, 内存曲线出现尖刺。7.5W 连接时无尖刺。

Server 8C16G, 带宽(上下行一样)为 450Mb/s

Client 4C8G * 4, 情况跟测试环境一样，带宽使用(上下行一样) 110Mb/s

一般做法，单机承载 5W 连接即可，规避宕机可能带来的大面积瘫痪。
按照目前的情况来算，2C2G机器即可完成 5W 负载。

## 坑
Client 库最好使用 Golang builtin net 库，gnet client 不能充分利用 CPU。