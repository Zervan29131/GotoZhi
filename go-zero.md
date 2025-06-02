# go-zero

## 简介

go-zero 是一个非常受欢迎的go语言微服务框架，截止到目前为止github上拥有高达28k的star。
它由国内大神Kevin Wan主导开发，它提供了许多开箱即用的功能，比如：限流、熔断、链路追踪、缓存、api参数自动校验、命令行代码生成等等。
复杂的事情都是从简单的事情开始的，麻雀虽小但是五脏俱全，先做一个最小、最简单的应用，然后吃透它，虽然不能做到完全掌握，但是至少能掌握和理解其中最重要核心的套路，以后即使面对比它更复杂的应用，也能快速掌握。
比如做一个论坛吧，架构简单，涉及的的表也比较少，类似于百度贴吧，任何用户都可以去发表言论，然后其它用户可以评论。

## 准备工作

- 本地安装Go，Go版本[go1.21.0](https://github.com/golang/go/releases/tag/go1.21.0)
- 本地安装 goctl，安装完后通过goctl --version查看是否安装成功。
- 本地安装protoc，protoc-gen-go
- 可访问的Mysql，Redis，Etcd，对版本不做要求
- IDE，VSCode
- 终端 iTerm2
- 调试工具 Postman
- 浏览器 Chrome

```shell
安装goctl
go install github.com/zeromicro/go-zero/tools/goctl@latest
下载超时可以使用如下代码进行配置代理
go env -w GOPROXY=https://goproxy.cn 
在终端中看到 zsh: command not found: goctl 的错误消息时
需要配置 GOPATH 和 GOROOT 环境变量
如果 goctl 在 bin 目录中，但仍然不能找到，需要将该目录添加到 PATH 中。
可以通过编辑 shell 配置文件（如 .zshrc 或 .bashrc）来实现
echo 'export PATH=$PATH:~/go/bin' >> ~/.zshrc
修改完后，运行以下命令以重新加载配置
source ~/.zshrc
安装了protoc编译器后还需要安装两个go相关的插件，protoc-gen-go、protoc-gen-go-grpc 用于生成go语言的grpc代码。
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
```

## 项目结构搭建

该项目是一个论坛系统，只需要用户（users）、帖子(posts)、评论(comments)三个模块。

我们把每个模块做为一个服务，每个服务我们分为api、rpc、model三个主要的部分。

```shell
mkdir forum && cd forum
go mod init forum

mkdir service
mkdir service/user 
mkdir service/user/{rpc,model,api}

mkdir service/post
mkdir service/post/{rpc,model,api}

mkdir service/comment
mkdir service/comment/{rpc,model,api}

mkdir common # 一个公用的目录 用于存放一些通用的代码
```

搭建后目录结构如下：


```go
tree
.
├── common
├── go.mod
└── service
    ├── comment
    │   ├── api
    │   ├── model
    │   └── rpc
    ├── post
    │   ├── api
    │   ├── model
    │   └── rpc
    └── user
        ├── api
        ├── model
        └── rpc
15 directories, 1 file
```

上面的每一个模块都是由rpc、model、api三个部分组成。
在传统的api服务中，只需要api去和model交互，但是在微服务架构中，会多一层rpc，由rpc去和model交互的，整体关系如下：
api -> rpc -> model

## model

首先使用user模块来演示

为了演示方便，我们使用mysql数据库，可以在本地先创建一个demo数据库database，然后创建一个users表，

```mysql
CREATE TABLE users (
    id bigint AUTO_INCREMENT,
    name varchar(255) NOT NULL DEFAULT '' COMMENT 'The username',
    password varchar(255) NOT NULL DEFAULT '' COMMENT 'The user password',
    mobile varchar(255) NOT NULL DEFAULT '' COMMENT 'The mobile phone number',
    gender char(10) NOT NULL DEFAULT 'male' COMMENT 'gender,male|female|unknown',
    nickname varchar(255) NULL DEFAULT '' COMMENT 'The nickname',
    type tinyint(1) NULL DEFAULT 0 COMMENT 'The user type, 0:normal,1:vip, for test golang keyword',
    create_at timestamp NULL DEFAULT CURRENT_TIMESTAMP,
    update_at timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE mobile_index (mobile),
    UNIQUE name_index (name),
    PRIMARY KEY (id)
) ENGINE = InnoDB COLLATE utf8mb4_general_ci COMMENT 'user table';
```

终端切换到/service/user/model目录下，在此目录下新建user.sql文件，然后将上述sql内容放进去。

命令行终端执行:

```shell
goctl model mysql ddl --src user.sql --dir .
```

看到Done则表示model代码生成成功。

在当前目录下生成三个文件：

vars.go 存放一些常量
usermodel.go  model初始化入口
usermodel_gen.go 数据库操作具体实现

这里着重关注下usermodel.go中的初始化入口。

```go
// 可以通过model.NewUserModel(sqlConn)来初始化model
func NewUserModel(conn sqlx.SqlConn) UserModel {
	return &customUserModel{
		defaultUserModel: newUserModel(conn),
	}
}
// 这里返回的UserModel是一个接口类型，这个接口需要实现Insert、FindOne、FindOneByMobile、FindOneByName、Update、Delete等方法

// 这些方法的实现是通过defaultUserModel这个结构体去实现的
```

除了上面的通过sql去生成model外，go-zero还可以通过当前的数据库中的表去生成model代码，使用如下命令：

```shell
goctl model mysql datasource --url="root:Passw0rd@tcp(127.0.0.1:3306)/demo" --table=users --dir=./
```

需要注意的是，执行goctl model命令并不会直接到我们本地的数据库创建表，因此我们需要手动到数据库中去新建表或增减字段。

虽然可以通过本地数据库直接生成model，但是为了别人拿到项目后能快速初始化表结构，还是建议在model层下放置完整的表sql文件。

## rpc

### rpc结构初始化

下面我们来开始创建rpc层，创建rpc首先需要创建proto文件，在/service/user/rpc目录下新建user.proto文件。

```protobuf
syntax = "proto3";

package user;
// go_package指定生成go包（也就是生成的.pb.go文件）的路径
// PS: 路径中要带/
// 在同级目录下执行 goctl rpc protoc user.proto --go_out=./ --go-grpc_out=./ --zrpc_out=. --client=true 生成
option go_package = "./user";

// 注册请求
message RegisterRequest {
  string Name = 1;
  string Mobile = 2;
  string Gender = 3;
	string Password = 4;
}

// 注册响应
message RegisterResponse {
  int64 Id = 1; // 注册完返回ID信息
	string Name = 2;
	string Mobile = 3;
	string Gender = 4;
}

// 登录请求
message LoginRequest {
  string Mobile = 1;
	string Password = 2;
}

// 登录响应
message LoginResponse {
	// 注意，登录后我们需要返回token的授权信息，但是这里并没有返回token
	// 是因为我们的token应该是api层去做的事，放在api层更合适
  int64 Id = 1;    
	string Name = 2;
	string Mobile = 3;
	string Gender = 4;
}

// 用户信息请求
message UserInfoRequest {
  int64 Id = 1;
}

// 用户信息响应
message UserInfoResponse {
  int64 Id = 1;
	string Name = 2;
	string Mobile = 3;
	string Gender = 4;
}

/*
  service 定义服务
  rpc 定义方法
  返回值类型
*/

// 这里命名为User 它生成客户端代码时，会生成一个userclient的目录
service User {
  rpc Register(RegisterRequest) returns (RegisterResponse);
	rpc Login(LoginRequest) returns (LoginResponse);
	rpc UserInfo(UserInfoRequest) returns (UserInfoResponse);
}
```

该文件是用于约定rpc请求和响应格式的，后续用这个文件来生成rpc代码。

从代码中可以看出，约定了三个rpc服务：注册、登录、用户信息。

一般而言，命名格式为xxRequest和xxResponse，其中xx为方法名。

而对server端的rpc方法，我们一般命名格式为xx，其中xx为方法名，字段名大写。

如果不清楚proto文件如何写，可以查看[书写 .proto 文件规范](https://www.zhangjiee.com/topic/grpc/write-proto-spec.html#org9cd8b2a)

然后在/service/user/rpc目录下执行命令生成rpc代码

```shell
//goctl rpc protoc user.proto --go_out=./ --go-grpc_out=./ --zrpc_out=. --client=true
//--client=true: 生成客户端代码。文件定义了一个数据库模型，主要作用是定义 UsersModel 接口和 customUsersModel 结构体，并提供与数据库表交互的方法，用于与数据库交互，而不涉及客户端与服务端的通信逻辑，不是用于 gRPC 或 ZeroRPC 通信。

goctl rpc protoc user.proto --go_out=./ --go-grpc_out=./ --zrpc_out=. 
```

看到Done则表示rpc代码生成成功了。

```shell
.
├── etc           # 配置文件存放 
│   └── user.yaml
├── internal      # 内部代码重点关注，我们需要改动代码基本都在这里
│   ├── config    # 配置层
│   │   └── config.go
│   ├── logic     # 业务逻辑层(业务具体实现位置)
│   │   ├── loginlogic.go
│   │   ├── registerlogic.go
│   │   └── userinfologic.go
│   ├── server   # 用于构建rpc server服务，转发logic处理
│   │   └── userserver.go
│   └── svc      # 服务上下文（里面封装了config)
│       └── servicecontext.go
├── user         # proto生成的中间代码，无需修改
│   ├── user.pb.go
│   └── user_grpc.pb.go
├── user.go      # rpc启动入口，rpc的main就在这里
├── user.proto  
└── userclient   # api层调用rpc入口
    └── user.go

9 directories, 12 files
```

生成文件后，需要先go mod tidy一下，拉取必要的依赖。
### rpc配置model关联

由于api与rpc交互，不直接与model交互，因此在生成rpc代码后，第一件事要做的就是配置数据库连接。

etc数据库配置在`service/user/rpc/etc/user.yaml`中配置mysql的连接地址。

```yaml
Name: user.rpc
ListenOn: 0.0.0.0:8080
Etcd:
  Hosts:
  - 127.0.0.1:2379
    Key: user.rpc
    
### 配置的是这部分
Mysql:
  DataSource: root:Passw0rd@tcp(127.0.0.1:3306)/demo?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai
```

config数据库配置在`service/user/rpc/internal/config/config.go`中修改，config结构体如下：

```go
type Config struct {
	zrpc.RpcServerConf

	// 这里引入mysql数据库配置 后面svc.ServiceContext中会用到
	// 这里的结构和etc配置文件中保持一致
	Mysql struct {
		DataSource string
	}
}
```
svc服务上下文配置在service/user/rpc/internal/svc/servicecontext.go中修改svc.ServiceContext结构体，引入model字段。

```go
type ServiceContext struct {
	Config    config.Config
	UserModel model.UsersModel // 添加model层关联
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:    c,
    // 添加UserModel字段关联
		UserModel: model.NewUsersModel(sqlx.NewMysql(c.Mysql.DataSource)),
	}
}
```

### rpc用户注册实现

有了前面的配置，实现逻辑就比较简单了，我们只需要在`service/user/rpc/internal/logic/registerlogic.go`中实现`RegisterLogic`方法即可。

```go
func (l *RegisterLogic) Register(in *user.RegisterRequest) (*user.RegisterResponse, error) {
	// 注册时 先去数据库查找下是否已经有这个人
	_, err := l.svcCtx.UserModel.FindOneByMobile(l.ctx, in.Mobile)
	if err == nil {
		return &user.RegisterResponse{}, status.Error(400, "该手机号已注册")
	}

	// 加这个判断是为了避免其它错误导致去创建
	if err == model.ErrNotFound {
		res, err := l.svcCtx.UserModel.Insert(l.ctx,
			&model.Users{
				Name:     in.Name,
				Mobile:   in.Mobile,
				Password: in.Password,
				Gender:   in.Gender,
			})

		if err != nil {
			return &user.RegisterResponse{}, status.Error(400, err.Error())
		}

		id, err := res.LastInsertId()
		if err != nil {
			return &user.RegisterResponse{}, status.Error(400, err.Error())
		}

		return &user.RegisterResponse{
			Id:     id,
			Name:   in.Name,
			Gender: in.Gender,
			Mobile: in.Mobile,
		}, nil
	}

	// 如果不是ErrNotFound，则说明数据库查询出错,这里报错500
	return &user.RegisterResponse{}, status.Error(500, err.Error())
}

```
上面的代码整体就是利用l变量-逻辑层去调用model层，做一些数据处理。接下来尝试是否能跑通注册功能。

在测试前面有两个地方需要注意：

1. 由于配置文件中是配置了etcd，因此需要先启动etcd服务。mac电脑可以通过brew 安装。
2. 通过grpcurl工具来测试rpc服务，默认是不能调试的，需要先开启dev模式，在`service/user/rpc/etc/user.yaml`中添加`Mode: dev`即可。

```yaml
Name: user.rpc
ListenOn: 0.0.0.0:8080
Mode: dev # 添加它

# ....

```

在`service/users/rpc`下启动服务。

```shell
go run user.go
Starting rpc server at 0.0.0.0:8080...
```

然后在再开一个终端，用于测试

```shell
$ grpcurl -plaintext 127.0.0.1:8080 list # 查看有哪些服务
grpc.health.v1.Health
grpc.reflection.v1.ServerReflection
grpc.reflection.v1alpha.ServerReflection
user.User
shell

$ grpcurl -plaintext 127.0.0.1:8080 list user.User # 查看某个服务下的方法
user.User.Login
user.User.Register
user.User.UserInfo
shell

# 这里需要注意字段名要和proto文件一致
$ grpcurl -plaintext -d '{"Name": "李四", "Mobile": "18200365766", "Password": "123456"}' 127.0.0.1:8080 user.User/Register # 调用服务的方法
{
  "Id": "3",
  "Name": "李四",
  "Mobile": "18200365766"
}
```

再去数据库查看，发现users表已经有数据了。

简单回顾下整个过程：

1. 根据proto文件生成rpc组织结构代码
2. rpc关联model修改（etc/config/svc 三个相关文件）
3. 到logic下编写业务逻辑

可以看出在框架搭建完成后，编写业务逻辑很快捷。

### rpc用户登录实现

有了前面的经验，我们照葫芦画瓢，修改用户登录业务代码`service/user/rpc/internal/logic/loginlogic.go`

```go
func (l *LoginLogic) Login(in *user.LoginRequest) (*user.LoginResponse, error) {
	// 先查用户
	u, err := l.svcCtx.UserModel.FindOneByMobile(l.ctx, in.Mobile)
	if err != nil {
		return &user.LoginResponse{}, status.Error(400, err.Error())
	}

	// 判断密码对么
	if u.Password != in.Password {
		return &user.LoginResponse{}, status.Error(400, "无效密码")
	}

	return &user.LoginResponse{
		Id:     u.Id,
		Name:   u.Name,
		Mobile: u.Mobile,
		Gender: u.Gender,
	}, nil
}
```

同理在终端进行测试效果如下：

```shell
$ grpcurl -plaintext -d '{"Mobile": "18200365766", "Password": "123456"}' 127.0.0.1:8080 user.User/Login
{
  "Id": "3",
  "Name": "李四",
  "Mobile": "18200365766"
}

grpcurl -plaintext -d '{"Mobile": "18200365766", "Password": "123"}' 127.0.0.1:8080 user.User/Login
ERROR:
  Code: Code(400)
  Message: 无效密码
```

### rpc用户信息实现

修改`service/user/rpc/internal/logic/userinfologic.go`文件

```go
func (l *UserInfoLogic) UserInfo(in *user.UserInfoRequest) (*user.UserInfoResponse, error) {
	u, err := l.svcCtx.UserModel.FindOne(l.ctx, in.Id)
	if err != nil {
		return &user.UserInfoResponse{}, status.Error(400, err.Error())
	}

	return &user.UserInfoResponse{
		Id:     u.Id,
		Name:   u.Name,
		Mobile: u.Gender,
		Gender: u.Gender,
	}, nil
}
```

终端测试

```shell
grpcurl -plaintext -d '{"Id": 3}' 127.0.0.1:8080 user.User/UserInfo
{
  "Id": "3",
  "Name": "李四"
}
```

### bcrypt密码

为了便于理解，密码数据库的存储和认证都是明文存储的。

但是实际生产中，一般都会对密码进行加密存储，防止数据库被攻击。

#### 密码公共包

使用bcrypt包时间密码的加密和对比。我们简单封装到`common`目录公共包中。

新建`common/bcryptx`目录，添加文件`common/bcrypt/bcryptx.go`文件，内容如下：

```go
package bcryptx

import "golang.org/x/crypto/bcrypt"

// password转hash
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashedPassword), err
}

// 验证密码
func ValidatePassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
```

#### 注册存密码hash

修改注册逻辑`service/user/rpc/internal/logic/registerlogic.go`

```go
func (l *RegisterLogic) Register(in *user.RegisterRequest) (*user.RegisterResponse, error) {
	// ...

	// 加这个判断是为了避免其它错误导致去创建
	if err == model.ErrNotFound {
		// 使用公用包的hash密码
		hashPass, err := bcryptx.HashPassword(in.Password)
		if err != nil {
			return &user.RegisterResponse{}, status.Error(400, err.Error())
		}

		res, err := l.svcCtx.UserModel.Insert(l.ctx,
			&model.Users{
				Name:     in.Name,
				Mobile:   in.Mobile,
				Password: hashPass, // 这里存hash密码
				Gender:   in.Gender,
			})

		// ...
	}

	// ...忽略
}
```

此时再此测试会发现，数据库里已经存的不是明文了。

#### 登录验证

登录部分需要同步修改成hash认证，修改文件`service/user/rpc/internal/logic/loginlogic.go`

```go
func (l *LoginLogic) Login(in *user.LoginRequest) (*user.LoginResponse, error) {
	// ...

	// 判断密码对么
	if err = bcryptx.ValidatePassword(u.Password, in.Password); err != nil {
		return &user.LoginResponse{}, status.Error(400, "无效密码")
	}

	// ...
}
```

## api实战

### api结构初始化

和rpc类似，api结构的初始化，我们也可以通过`goctl`命令生成，它的生成也是通过定义`.api`文件实现的，这是一个专属语言的格式，我们按照格式写就行。

切换到`service/user/api`目录下，添加`user.api`文件，内容如下：

```go
// api路径下执行 goctl api go -api ./user.api -dir ./
type (
	// 注册请求
	RegisterRequest {
		Name     string `json:"name"`
		Mobile   string `json:"mobile"`
		Gender   string `json:"gender"`
		Password string `json:"password"`
	}
	// 注册响应
	RegisterResponse {
		ID     int64  `json:"id"`
		Name   string `json:"name"`
		Mobile string `json:"mobile"`
		Gender string `json:"gender"`
	}
)

// api定义的地方
service user {
	@handler Register // 注册接口请求的方法名
	post /api/user/register (RegisterRequest) returns (RegisterResponse)
}
```

如果不清楚api定义的格式，可以参考[api文档](https://link.juejin.cn?target=https%3A%2F%2Fgo-zero.dev%2Fdocs%2Ftasks%2Fdsl%2Fapi) 终端执行api代码生成命令`goctl api go -api ./user.api -dir ./`

```shell
.
├── etc          # 配置文件存放位置
│   └── user.yaml
├── internal     # 内部代码位置 重点关注
│   ├── config   # 配置层
│   │   └── config.go
│   ├── handler  # 特殊（只有api层有）作用是路由导向，到logic层
│   │   ├── registerhandler.go
│   │   └── routes.go
│   ├── logic    # 逻辑层（封装了服务上下文svc） 业务主要处理层
│   │   └── registerlogic.go
│   ├── svc      # 服务上下文（封装了config)
│   │   └── servicecontext.go
│   └── types    # 存放请求响应结构体（一般无需修改）
│       └── types.go
├── user.api     # 定义api的专属文件
└── user.go      # api启动入口

8 directories, 9 files
```

基本和`rpc`一样，只是多了一个特殊的hander层用于路由导向到`logic`，功能类比与rpc中的server。

### api配置rpc关联

api不直接与model通信，它是和rpc通信，所以需要将api和rpc做关联，这一步可以类比与rpc与model关联。

#### 配置etc文件

在`service/user/api/etc/user.yaml`中添加

```yaml
yaml

 代码解读
复制代码#...省略

# 一个rpc一个配置项
# rpc通过etcd来做服务发现和注册
UserRpc:
  Etcd:
    Hosts:
    - 127.0.0.1:2379 # 注意这里是etcd的地址而非rpc服务的地址
    Key: user.rpc
```

#### 配置config

修改`service/user/api/internal/config/config.go`文件

```go
go

 代码解读
复制代码type Config struct {
	rest.RestConf

	// 这里直接定义一个字段就行 在config初始化时，会自动将etc中rpc配置加载到config中
	// zrpc.RpcClientConf 是一个结构体,在svc中初始化上下文时使用
	UserRpc zrpc.RpcClientConf
}
```

#### 配置svc

修改`service/user/api/internal/svc/servicecontext.go`

```go
go

 代码解读
复制代码type ServiceContext struct {
	Config config.Config

	UserRpc userclient.User // 关联user rpc,userclient是rpc中提供的客户端包
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:  c,
    // 这里c.UserRpc取config中的配置此时已从配置中载入
		UserRpc: userclient.NewUser(zrpc.MustNewClient(c.UserRpc)),
	}
}
```

配置完成，下面开始实现接口。

### 注册api实现

直接修改文件注册logic文件`service/user/api/internal/logic/registerlogic.go`

```go
func (l *RegisterLogic) Register(req *types.RegisterRequest) (resp *types.RegisterResponse, err error) {
	// 注意这里是直接使用的userclient中的结构体组装的，并无单独的工厂方法
	res, err := l.svcCtx.UserRpc.Register(l.ctx, &userclient.RegisterRequest{
		Name:     req.Name,
		Mobile:   req.Mobile,
		Gender:   req.Gender,
		Password: req.Password,
	})

	if err != nil {
		return &types.RegisterResponse{}, err
	}

	return &types.RegisterResponse{
		ID:     res.Id,
		Name:   res.Name,
		Mobile: res.Mobile,
		Gender: res.Gender,
	}, nil
}
```

现在我们来测试下，由于api依赖rpc所以我们需要分别开启rpc服务和api服务，然后再测试接口。

切换到`service/user/rpc`目录下,开启rpc服务

```shell
go run user.go
Starting rpc server at 0.0.0.0:8080...
```

切换到`service/user/api`下,开启api服务

```shell
go run user.go
Starting server at 0.0.0.0:8888...
```

curl测试

```shell
curl --location 'http://localhost:8888/api/user/register' \
--header 'Content-Type: application/json' \
--data '{
    "name": "keke",
    "mobile": "17655434667",
    "password": "ksdafsda",
    "gender": "male"
}'

{
    "id": 7,
    "name": "keke",
    "mobile": "17655434667",
    "gender": "male"
}
```

### jwt鉴权

在开始实现登录前，我们需要有鉴权机制。

#### api配置jwt

`go-zero`是支持jwt鉴权的，直接在`service/user/api/user.api`中添加`@service`块就行。

```go
// ...忽略

// 其下的所有service都会使用jwt鉴权,注意这里是server不是service
@server (
	jwt: Auth
)
// 具体服务先留着
service User {}
```

#### token签发

jwt需要签发token，我们使用`jwt`包实现,这部分属于公用逻辑，我们封装到`common`目录下。

新建`common/jwtx`目录，下面添加文件`common/jwtx/jwt.go`

```go
package jwtx

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// GenToken 生成JWT令牌
// 参数:
//
//	uid: 用户唯一标识
//	exp: 令牌过期时间
//	signKey: 签名密钥
//
// 返回值:
//
//	生成的令牌字符串
//	错误对象，如果生成令牌过程中出现错误
func GenToken(uid int64, exp time.Time, signKey string) (string, error) {
	claims := jwt.MapClaims{
		"uid": uid, // 用户id
		"exp": exp.Unix(),
		"iat": time.Now().Unix(), // 签发时间
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(signKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
```

#### etc配置jwt

签名时间、过期时间 在`service/user/api/etc/user.yaml`中添加下面信息

```yaml
# 用于jwt授权
Auth:
  AccessSecret: "abcdefg22434" # 签名密钥 注意这里要求长度不少于8
  AccessExpire: 3600      # 单位秒 默认1小时
```

#### config添加Auth配置

修改`service/user/api/internal/config/config.go`文件

```go
type Config struct {
	rest.RestConf

	// 这里直接定义一个字段就行 在config初始化时，会自动将etc中rpc配置加载到config中
	// zrpc.RpcClientConf 是一个结构体,在svc中初始化上下文时使用
	UserRpc zrpc.RpcClientConf

	// 授权配置信息
	Auth struct {
		AccessSecret string
		AccessExpire int64
	}
}
```

### 登录api实现

有了前面的jwt铺垫，我们实现登录api就很容易了,我们在业务上只需要，在认证成功后，返回token，后过期时间即可。

#### api文件添加接口信息

修改`service/user/api/user.api`文件

```go
type (
	// ...

  // 登录请求
	LoginRequest {
		Mobile   string `json:"mobile"`
		Password string `json:"password"`
	}
	// 登录响应
	LoginResponse {
		Token   string `json:"token"`
		Expired int64  `json:"expired"`
	}
)

// api定义的地方
service user {
	// ....

	@handler Login
	post /api/user/login (LoginRequest) returns (LoginResponse)
}
```

执行`goctl api go -api user.api -dir .`自动生成代码。

#### logic实现

修改`service/user/api/internal/logic/loginlogic.go`文件

```go
func (l *LoginLogic) Login(req *types.LoginRequest) (resp *types.LoginResponse, err error) {
	res, err := l.svcCtx.UserRpc.Login(l.ctx, &user.LoginRequest{
		Mobile:   req.Mobile,
		Password: req.Password,
	})

	if err != nil {
		return nil, err
	}
	// 到期时间
	expireTime := time.Now().Add(time.Duration(l.svcCtx.Config.Auth.AccessExpire) * time.Second)
	secret := l.svcCtx.Config.Auth.AccessSecret
	// 生成签名token
	accessToken, err := jwtx.GenToken(res.Id, expireTime, secret)

	if err != nil {
		return nil, err
	}

	return &types.LoginResponse{
		Token:   accessToken,
		Expired: expireTime.Unix(),
	}, nil
}
```

启动rpc、以及api服务终端测试

```shell
curl --location 'http://localhost:8888/api/user/login' \
--header 'Content-Type: application/json' \
--data '{
    "mobile": "18200365866",
    "password": "12312"
}'
{"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MjA5NDEzMDcsImlhdCI6MTcyMDkzNzcwNywidWlkIjo1fQ.deZRXcuyydg3DgpHURXD-SZDJ2ct3gZvLpWGe1e0rGY","expired":1720941307}
```

### 用户信息api实现

在登录的情况下，由于请求时带上了jwt token，而token中已经存放了用户id，所以可以直接通过id获取用户信息。而无需其它参数。

#### api文件添加接口信息

修改`service/user/api/user.api`文件

```go
type (
	// ...

	// 用户信息请求 由于无需参数所以不写
	// 用户信息响应
	UserInfoResponse {
		ID     int64  `json:"id"`
		Name   string `json:"name"`
		Mobile string `json:"mobile"`
		Gender string `json:"gender"`
	}
)

// 其下的所有service都会使用jwt鉴权,注意这里是server不是service
@server (
	jwt: Auth
)
service user {
	@handler UserInfo
	// 这里没有请求参数哦
	get /api/user/info returns (UserInfoResponse)
}
```

执行`goctl api go -api user.api -dir .`自动生成代码。

#### logic实现

修改`service/user/api/internal/logic/userinfologic.go`文件

```go
func (l *UserInfoLogic) UserInfo() (resp *types.UserInfoResponse, err error) {
	// 注意这里ctx的uid是框架自动帮我们解析出来
	// uid是签发授权时存的user id,这里取出来的是interface{}类型
	uid, err := l.ctx.Value("uid").(json.Number).Int64()
	if err != nil {
		return nil, err
	}
	user, err := l.svcCtx.UserRpc.UserInfo(l.ctx, &userclient.UserInfoRequest{
		Id: uid,
	})
	if err != nil {
		return nil, err
	}

	return &types.UserInfoResponse{
		ID:     user.Id,
		Name:   user.Name,
		Mobile: user.Mobile,
		Gender: user.Gender,
	}, nil
}
```

启动服务执行测试

```shell
curl --location 'http://localhost:8888/api/user/info' \
--header 'Authorization: Bearer eyJhbGciOiJIUzI1NiJ9.eyJ1aWQiOjV9._uaaeq2_bzyJomQLVg9-ZH7kxdpWZ565eum6ZJYgcRI'
```

这里的token可以从登录接口获取,也可以根据工具生成一个token值。

## 总结

1. 无论是rpc还是api,我们第一步都是去写文档（.api/.proto），约定接口和参数，然后自动生成代码
2. 做配置，rpc要关联model，api要关联rpc,都是要去etc文件配置、config配置、svc配置
3. 配置完成后，到逻辑层编写业务逻辑



