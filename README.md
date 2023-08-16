



<p align="center">
  <a href="https://github.com/wgpsec/ENScan_GO">
    <img src="README/logo.png" alt="Logo" width="80" height="80">
  </a>
  <h3 align="center">ENScan Go</h3>
  <p align="center">
    剑指HW/SRC，解决在HW/SRC场景下遇到的各种针对国内企业信息收集难题
    <br />
          <br />
<a href="https://github.com/wgpsec/ENScan_GO/stargazers"><img alt="GitHub stars" src="https://img.shields.io/github/stars/wgpsec/ENScan_GO"/></a>
<a href="https://github.com/wgpsec/ENScan_GO/releases"><img alt="GitHub releases" src="https://img.shields.io/github/release/wgpsec/ENScan_GO"/></a>
<a href="https://github.com/wgpsec/ENScan_GO/blob/main/LICENSE"><img alt="License" src="https://img.shields.io/badge/License-Apache%202.0-blue.svg"/></a>
<a href="https://github.com/wgpsec/ENScan_GO/releases"><img alt="Downloads" src="https://img.shields.io/github/downloads/wgpsec/ENScan_GO/total?color=brightgreen"/></a>
<a href="https://goreportcard.com/report/github.com/wgpsec/ENScan_GO"><img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/wgpsec/ENScan_GO"/></a>
<a href="https://twitter.com/wgpsec"><img alt="Twitter" src="https://img.shields.io/twitter/follow/wgpsec?label=Followers&style=social" /></a>
<br>
<br>
<a href="https://github.com/wgpsec/ENScan_GO/discussions"><strong>探索更多Tricks »</strong></a>
      <br/>
    <br />
    <a href="https://github.com/wgpsec/ENScan_GO/releases">下载程序</a>
    ·
    <a href="https://github.com/wgpsec/ENScan_GO/issues">反馈Bug</a>
    ·
    <a href="https://github.com/wgpsec/ENScan_GO/discussions">提交需求</a>
  </p>



## 功能列表

**使用程序可能导致账号被封，仅用于信息收集用途，请勿用于非法用途**

**若该程序影响或侵犯到您的权益，请与我们联系**

![ENScanGo](README/ENScanGo.png)

 - 使用支持以下API，并支持合并数据导出
    - 爱企查 (未登陆信息带*)
    - 天眼查
    - 阿拉丁
    - 酷安市场
    - 七麦数据
    - 站长之家
    - veryvp
 - 查询信息
    - ICP备案
    - APP
    - 微博
    - 微信公众号
    - 控股公司
    - 供应商
    - 客户信息
    - 小程序
    - 控股X的公司的以上所有信息
    - ...
 - 通过APK市场收集使用信息



## 使用指南
请直接前往RELEASE下载编译好的文件使用

ENScanGo在第一次使用时需要使用 -v 命令 生成配置文件信息

**遇到问题请加上参数 --debug 提issue**

如果查询不出来目标网站信息，可以都挂上BURP代理进行查询

### 登陆配置

**AQC**

请注意获取COOKIE域名，默认查询为aiqicha.baidu.com，请勿使用 aiqicha.com

Cookie信息请勿直接 `document.cookie`，可能因为http-only 选项无法复制全导致登陆失败

![image-20221028223835307](README/image-20221028223835307.png)

**阿拉丁使用的是TOKEN**

点击任意请求，拿到请求里的TOKEN

![image-20230709132223242](./README/image-20230709132223242.png)

**TYC tycid**

配置COOKIE后配置tycid

![image-20230722194839975](./README/image-20230722194839975.png)

### 快速使用

**默认公司信息** (网站备案, 微博, 微信公众号, app)

```
./enscan -n 小米
```

**对外投资占股100%的公司 获取孙公司（深度2）**

```
./enscan -n 小米 -invest 100 -deep 2
```

**组合筛选** 

大于51%控股公司、供应商、分支机构，只要ICP备案信息，并且批量获取邮箱信息

```
./enscan -n 小米 -field icp --hold --supplier --branch --email 
```

**使用不同渠道** 

使用天眼查数据源（或可设定为 all 组合多个数据源）

```
./enscan -n 小米 -type tyc
```

使用多数据源一起收集（暂不支持多渠道+筛选）

```
./enscan -n 小米 -type aqc,chinaz,qimai,coolapk
```

### 选项说明

#### **field 获取字段**

使用参数 `field`指定需要查询的信息，可指定多参数一起查询，方便快速收集

```
-n 小米 -field icp,app
```

支持以下参数

- `icp` 网站备案信息
- `weibo` 微博
- `wechat` 微信公众号
- `app` 应用信息
- `job` 招聘信息
- `wx_app` 微信小程序
- `copyright` 软件著作权
- `supplier` 供应商信息（通过招标书确定）

#### **type 获取字段**

使用参数 `type`可以指定需要API数据源

```
-n 小米 -type tyc
```

- `aqc`   爱企查
- `tyc`   天眼查
- `all`   全部查询
- `aldzs` 阿拉丁 （仅小程序）
- `coolapk` 酷安市场 （仅APP）
- `qimai` 七麦数据（仅APP）
- `chinaz` 站长之家（仅ICP备案）

#### 完整参数

| 参数              | 样例           | 说明                                   |
| ----------------- | -------------- | -------------------------------------- |
| -n                | 小米           | 关键词                                 |
| -i                | 29453261288626 | 公司PID（自动识别类型）                |
| -f                | file.txt       | 批量查询，文本按行分隔（可选PID模式）  |
| -type             | aqc            | API类型                                |
| -o                |                | 结果输出的文件夹位置(可选)             |
| -is-merge         |                | 合并导出                               |
| -invest           |                | 投资比例                               |
| -field            | icp            | 获取字段信息                           |
| -deep             | 1              | 递归搜索n层公司                        |
| -hold             |                | 是否查询控股公司                       |
| -supplier         |                | 是否查询供应商信息                     |
| -branch           |                | 查询分支机构（分公司）信息             |
| -is-branch        |                | 深度查询分支机构信息（数量巨大）       |
| -api              |                | 是否API模式                            |
| -client           |                | 客户端模式通道                         |
| -debug            |                | 是否显示debug详细信息                  |
| -is-show          |                | 是否展示信息输出                       |
| -uncertain-invest |                | 包括未公示投资公司（无法确定占股比例） |
| -is-group         |                | 查询关键词为集团                       |
| -is-pid           |                | 批量查询文件是否为公司PID              |
| -delay            |                | 每个请求延迟（S）-1为随机延迟1-5S      |
| -proxy            |                | 设置代理                               |
| -timeout          |                | 每个请求默认1（分钟）超时              |
| -no-merge         |                | 批量查询【取消】合并导出               |
| -v                |                | 版本信息                               |
| -email            |                | 获取email信息                          |

### API模式

ENScanGo可使用API模式进行分布式部署，搭建API服务构建资产处理

使用`rmq`作为任务队列组件，在redis存入任务信息，可分部署部署。数据可存储至mongodb进行调用分析。

**api调用效果（前端开发中）**

可使用 https://enscan.wgpsec.org/api/info 体验

![image-20221028231744940](README/image-20221028231744940.png)

![image-20221028231815437](README/image-20221028231815437.png)

![image-20221028231831102](README/image-20221028231831102.png)

![image-20221028232013627](README/image-20221028232013627.png)

#### API说明

在获取信息的内容在数据库内不存在时，会自动添加队列任务收集相关信息

**状态信息**

返回当前系统状态信息

```
GET /status
```

**获取信息**

```
GET /api/info?search=小米&invest=100&branch=true
```

| 参数      | 参数                 | 说明                       |
| --------- | -------------------- | -------------------------- |
| orgname   | 文本                 | 完整公司名称（二选一）     |
| search    | 文本                 | 模糊匹配公司名称（二选一） |
| type      | 文本，与命令参数一致 | 数据源                     |
| field     | 文本，与命令参数一致 | 筛选指定信息               |
| duplicate | true                 | 加上参数去重               |
| depth     | 数字                 | 爬取几层公司 如 2 为孙公司 |
| invest    | 数字                 | 筛选投资比例               |
| holds     | true                 | 筛选控股公司               |
| supplier  | true                 | 筛选供应商信息             |
| branch    | true                 | 筛选分支信息               |
| output    | true                 | 为true导出excel表格        |

**股权穿透**

返回投资、股东关联信息

```
GET /api/stockchart
```

| 参数      | 参数 | 说明             |
| --------- | ---- | ---------------- |
| orgname   | 文本 | 完整公司名称     |
| search    | 文本 | 模糊匹配公司信息 |
| duplicate | true | 是否去重         |

**添加获取任务**

添加新的任务信息

```
POST /api/info
```

| 参数      | 参数                 | 说明         |
| --------- | -------------------- | ------------ |
| orgname   | 文本                 | 完整公司名称 |
| update    | true                 | 是否更新     |
| type      | 文本，与命令参数一致 | 数据源       |
| invest_rd | true                 | 不确定投资   |
| branch    | true                 | 获取分支信息 |
| field     | 文本，与命令参数一致 | 筛选字段     |
| invest    | 数字                 | 投资信息     |

#### 启动部署

**golang 版本依赖**
```
go >= 1.17
```

然后我们需要对配置文件进行修改，加入数据库连接信息，与common同级加入以下配置

```yaml
api:
  server: "127.0.0.1" # redis地址
  mongodb: "mongodb://user:pass@127.0.0.1:27017" # mongodb 连接信息
  redis: "redis_password" # redis 密码
  port: "8080" # 启动API端口
```

mongodb数据库初始化配置

```
db.createCollection("app");
db.createCollection("branch");
db.createCollection("copyright");
db.createCollection("enterprise_info");
db.getCollection("enterprise_info").createIndex({
    name: "hashed"
}, {
    name: "name"
});
db.createCollection("holds");
db.createCollection("icp");
db.createCollection("infos");
db.createCollection("invest");
db.createCollection("job");
db.createCollection("partner");
db.createCollection("supplier");
db.createCollection("wechat");
db.createCollection("weibo");
db.createCollection("wx_app");
```

**API模式**

启动API模式将在配置端口监听，并启动api服务，可通过api服务进行调用读取数据

在此场景下无需配置cookie信息

```
./enscan --api
```

**客户端模式**

可作为部署节点，读取任务队列自动读取信息写入数据库

```
./enscan --client
```







##  交流&反馈

关注公众号 `WgpSec狼组安全团队` 回复`加群`一起交流~

![](https://assets.wgpsec.org/www/images/wechat.png)

[![Stargazers over time](https://starchart.cc/wgpsec/ENScan_GO.svg)](https://starchart.cc/wgpsec/ENScan_GO)



## 404星链计划

<img src="https://github.com/knownsec/404StarLink/raw/master/Images/logo.png" width="30%">

ENScanGo 现已加入 [404星链计划](https://github.com/knownsec/404StarLink)



##  JetBrains OS licenses
``ENScanGo`` had been being developed with `GoLand` IDE under the **free JetBrains Open Source license(s)** granted by JetBrains s.r.o., hence I would like to express my thanks here.

<a href="https://www.jetbrains.com/?from=wgpsec" target="_blank"><img src="https://raw.githubusercontent.com/wgpsec/.github/master/jetbrains/jetbrains-variant-4.png" width="256" align="middle"/></a>
