
# ENScan_GO

ENScanGo 是现有开源项目 [ENScan](https://github.com/wgpsec/ENScan) 的升级版本，为避免滥用导致API失效，目前仅开源部分代码供参考！

![](https://shields.io/badge/Go-1.17-green?logo=go)



<p align="center">
  <a href="https://github.com/wgpsec/ENScan_GO">
    <img src="README/logo.png" alt="Logo" width="80" height="80">
  </a>

  <h3 align="center">ENScan的Go版本实现</h3>
  <p align="center">
    解决遇到的各种针对国内企业信息收集难题
    <br />
    <a href="https://github.com/wgpsec/ENScan_GO"><strong>探索更多Tricks »</strong></a>
    <br />
    <br />
    <a href="https://github.com/wgpsec/ENScan_GO/releases">下载可执行文件</a>
    ·
    <a href="https://github.com/wgpsec/ENScan_GO/issues">反馈Bug</a>
    ·
    <a href="https://github.com/wgpsec/ENScan_GO/issues">提交需求</a>
  </p>




### 功能列表

 - 企业基本信息收集
    - ICP备案
    - APP
    - 微博
    - 微信公众号
    - 子公司
    - 供应商
    - ...
 - 小程序信息收集（包括同主体下的小程序）
 - 安卓App（apk）收集

### 使用指南

命令行参数如下
```
  -branch 是否拿到分支机构详细信息，为了获取邮箱和人名信息等
  -c string Cookie信息
  -f string 包含公司ID的文件
  -flags string 获取哪些字段信息
  -i string 公司ID
  -invest-num int 筛选投资比例，默认0为不筛选
  -invest-rd 是否选出不清楚投资比例的（出现误报较高）
  -n string 公司名称
  -o string 结果输出的文件(可选)
  -type string 选择收集渠道 (default "a")
  -v	版本信息
```

