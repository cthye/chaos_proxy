nessaj_proxy用于在内网环境下中转控制端的请求给无法直接访问的agent机器。

## development

### 证书生成

proxy涉及两套key pair，记为KP1(`SenderKeyPair`)和KP2(`ReceiverKeyPair`)，则proxy需要知道KP1的密钥和KP2的公钥：

- KP1用于自身向agent发送指令时生成agent需要的JWT token，所以需要private key
- KP2用于接收管理端的请求时进行权限验证，作用和agent启动时指定的公钥一样。

KP1和KP2统一都为prime256v1的椭圆曲线key pair，本地开发测试时可以通过如下方式生成(一对):

```sh
## sender key pair
openssl ecparam -out send_key.pem -name prime256v1 -genkey
# no need to generate public key for sender key
# openssl ec -in send_key.pem -pubout -out send_pub.pem

## receiver key pair
openssl ecparam -out recv_key.pem -name prime256v1 -genkey
openssl ec -in recv_key.pem -pubout -out recv_pub.pem
```

### 相关package

- 配置读取使用[viper](https://github.com/spf13/viper)库
- 命令行参数解析暂定使用[pflag](https://github.com/spf13/pflag)
- http服务使用[gin](https://github.com/gin-gonic/gin)
- agent信息持久化存储方案[badger](https://github.com/dgraph-io/badger)

## 鉴权

### 接受来自管理端的请求

程序启动时解析公钥(KP2)，并对所有请求进行签名验证(JWT)。作为请求方需要按如下方式添加签名信息：

1. 准备private key
2. 构造签名的内容：`{"iat": 1594092544, "exp": 1594092559}`，其中`iat`代表`issue at`，`exp`代表过期时间，建议设置两者相差不超过30秒。详情参考JWT的[RFC](https://tools.ietf.org/html/rfc7519#page-9)
3. 使用private key对上面内容按照JWT规范签名，算法使用`ES256` (ECDSA P-256 curve, SHA-256 hash algo)，得到`token`
4. Header头带上签名`Authorization: Bearer <token>`

### agent信息持久化存储

proxy需要在database目录下生成data目录作为badger存放数据的目录

python

```python
import time
import requests
import jwt

key = open('./ec_key.pem').read()

url = ''
now = int(time.time())
auth = jwt.encode({'iat': now, 'exp': now + 10}, key, 'ES256').decode()
print(auth)
r = requests.get(url, headers={'Authorization': f'Bearer {auth}'})
print(r.status_code)
```

### 转发请求到agent端

TODO
