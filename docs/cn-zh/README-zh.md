<h1 align="center">headscale-panel</h1>

<div align="center">
本项目时在headscale-panel项目基础上进行开发的，采用Go + Vue开发的管理系统脚手架, 前后端分离, 仅包含项目开发的必需部分, 基于角色的访问控制(RBAC), 分包合理, 精简易于扩展。 后端Go包含了gin、 gorm、 jwt和casbin等的使用, 前端Vue基于vue-element-admin开发: https://github.com/QianheYu/headscale-panel-ui.git
</div>

## 特性

- `Gin` 一个类似于martini但拥有更好性能的API框架, 由于使用了httprouter, 速度提高了近40倍
- `Postgres` 采用的是Postgres数据库
- `Jwt` 使用JWT轻量级认证, 并提供活跃用户Token刷新功能
- `Casbin` Casbin是一个强大的、高效的开源访问控制框架，其权限管理机制支持多种访问控制模型
- `Gorm` 采用Gorm 2.0版本开发, 包含一对多、多对多、事务等操作
- `Validator` 使用validator v10做参数校验, 严密校验前端传入参数
- `Lumberjack` 设置日志文件大小、保存数量、保存时间和压缩等
- `Viper` Go应用程序的完整配置解决方案, 支持配置热更新
- `GoFunk` 包含大量的Slice操作方法的工具包

## 中间件

- `AuthMiddleware` 权限认证中间件 -- 处理登录、登出、无状态token校验
- `RateLimitMiddleware` 基于令牌桶的限流中间件 -- 限制用户的请求次数
- `OperationLogMiddleware` 操作日志中间件 -- 记录所有用户操作
- `CORSMiddleware` -- 跨域中间件 -- 解决跨域问题
- `CasbinMiddleware` 访问控制中间件 -- 基于Casbin RBAC, 精细控制接口访问

## 安装方法
### 安装数据库
你可以采用直接部署或Docker的方式安装数据库，此处仅介绍使用Docker安装。如你已经安装了Postgres数据库请跳过这一步
复制 `deploy/postgres/docker-compose.yml`文件到你想要存放数据库数据文件的地方,根据文件中的提示修改你想修改的内容，然后执行以下命令
```shell
docker-compose up -d
```
### 准备配置文件
根据配置文件示例修改配置文件
### 安装headscale-panel
你可以选择如下两种部署方式选择合适的方法
#### Docker部署(推荐)
拉取docker镜像
```shell
docker pull yqhdocker/headscale-panel:latest
```
初始化数据库
```shell
docker run -it --net=host -v /your/config/path:/etc/headscale-panel --name=headscale-panel yqhdocker/headscale-panel:latest headscale-panel --init
```
> 注意: 如果你在安装headscale-panel前已经由headscale初始化了数据库请使用项目中的脚本修改数据库 [说明](./docs/InitDatabase.md)

运行headsale-panel
```shell
docker start headscale-panel
```

#### 直接部署
在Release中下载可执行文件并放到`/usr/local/bin`中
将配置文件放到`/etc/headscal-panel`目录中或运行时使用 `-c` 参数指定配置文件
初始化数据库
```shell
headscale-panel --init
```

设置服务管理器
```shell
# 以ubuntu server为例, 复制systemd/headscale-panel.service文件到/etc/systemd/system
systemctl daemon-reload
systemctl start headscale-panel
# 设置随系统启动
systemctl enable headscale-panel
```

## 项目截图

![登录](https://github.com/gnimli/go-web-mini-ui/blob/main/src/assets/GithubImages/login.PNG)
![用户管理](https://github.com/gnimli/go-web-mini-ui/blob/main/src/assets/GithubImages/user.PNG)
![角色管理](https://github.com/gnimli/go-web-mini-ui/blob/main/src/assets/GithubImages/role.PNG)
![角色权限](https://github.com/gnimli/go-web-mini-ui/blob/main/src/assets/GithubImages/rolePermission.PNG)
![菜单管理](https://github.com/gnimli/go-web-mini-ui/blob/main/src/assets/GithubImages/menu.PNG)
![API管理](https://github.com/gnimli/go-web-mini-ui/blob/main/src/assets/GithubImages/api.PNG)
![设备管理](./docs/images/machine.png)
![子网管理](./docs/images/subroute.png)
![PreAuthKey](./docs/images/preauthkey.png)
![Headscale设置](./docs/images/headscaleconfig.png)
## 项目结构概览

```
├─common # casbin postgres validator 等公共资源
├─config # viper读取配置
├─controller # controller层，响应路由请求的方法
├─dto # 返回给前端的数据结构
├─log # 日志模块zap
├─middleware # 中间件
├─model # 结构体模型
├─repository # 数据库操作
├─response # 常用返回封装，如Success、Fail
├─routes # 所有路由
├─util # 工具方法
├─task # 管理和连接headscale
└─vo # 接收前端请求的数据结构

```
## 前端Vue项目
[https://github.com/QianheYu/headscale-panel-ui.git](https://github.com/QianheYu/headscale-panel-ui.git)

## MIT License

    Copyright (c) 2023 QianheYu

