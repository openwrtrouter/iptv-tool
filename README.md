# iptv-tool

[![GitHub Release](https://img.shields.io/github/v/release/super321/iptv-tool?logo=github)](https://github.com/super321/iptv-tool/releases/latest)
[![GitHub Downloads (all assets, all releases)](https://img.shields.io/github/downloads/super321/iptv-tool/total?logo=github)](https://github.com/super321/iptv-tool/releases/latest)

IPTV工具，功能列表如下：

* 自动更新频道列表和EPG信息。
* 提供m3u、txt和pls格式直播源在线接口。
    * 支持频道黑名单过滤、频道分组以及频道台标配置
    * 支持m3u的catchup回看参数配置
* 提供EPG在线接口，支持xmltv和json两种格式。

## 配置说明

详细说明参见配置文件[config.yml](./config.yml)

## 使用介绍

将config.yml配置文件与工具放在一起，然后运行工具即可，具体运行命令如下：

* 根据某次抓包获取的Authenticator反向破解key

```
./iptv key -a xxxxx
```

说明：-a后面指定Authenticator，待运行完毕后会在当前目录下生成key.txt文件，其中可能找到很多key，任意一个均可使用(文件中Find
Key后面的即是)。
更多参数说明可通过命令`./iptv key -h`查看。

* 直接生成m3u直播源文件

```
./iptv channel -f m3u -u http://192.168.3.1:4022
```

说明：运行完毕后会在当前目录下生成iptv.m3u文件，通过-u参数指定软路由的udpxy的http地址。
更多参数说明可通过命令`./iptv channel -h`查看。

* 启动HTTP服务，提供在线m3u和epg接口：

```
./iptv serve -i 24h -p 8088 -u http://192.168.3.1:4022
```

或

```
./iptv serve -i 24h -p 8088 -u inner=http://192.168.3.1:4022
```

说明：-i指定频道和EPG更新间隔时间，-p指定启动的http服务的端口，-u指定udpxy的http地址。
更多参数说明可通过命令`./iptv serve -h`查看。

## HTTP API

* [m3u格式直播源](#m3u格式直播源)
* [txt格式直播源](#txt格式直播源)
* [pls格式直播源](#pls格式直播源)
* [json格式EPG](#json格式EPG)
* [xmltv格式EPG](#xmltv格式EPG)
* [xmltv格式EPG（gzip压缩）](#xmltv格式epggzip压缩)

### m3u格式直播源

```
http://IP:PORT/channel/m3u?csFormat={format}&multiFirst={multiFirst}&udpxy={udpxy}
```

#### 参数说明

* csFormat：可指定回看catchup-source的请求格式，支持通过配置文件[config.yml](./config.yml)中的`catchup.sources`进行自定义配置。
  **非必填，缺省为其中任意一个**。

  > 例如，若config.yml部分内容为：<br/>
  > ```
  > catchup:
  >   sources:
  >     diyp: "playseek=${(b)yyyyMMddHHmmss}-${(e)yyyyMMddHHmmss}"
  >     kodi: "playseek={utc:YmdHMS}-{utcend:YmdHMS}"
  > ```
  > * `/channel/m3u?csFormat=diyp`则使用`playseek=${(b)yyyyMMddHHmmss}-${(e)yyyyMMddHHmmss}`。
  > * `/channel/m3u?csFormat=kodi`则使用`playseek={utc:YmdHMS}-{utcend:YmdHMS}`。
  > * `/channel/m3u?csFormat=notexist`若指定的名称不存在，则不生成catchup相关内容。

  若未填写配置文件[config.yml](./config.yml)中的`catchup.sources`内容，则缺省使用以下内容：

| 值 | 是否缺省 | 说明                                                  |
|---|------|-----------------------------------------------------|
| 0 | 是    | `playseek=${(b)yyyyMMddHHmmss}-${(e)yyyyMMddHHmmss}` |
| 1 | 否    | `playseek={utc:YmdHMS}-{utcend:YmdHMS}`             |

* multiFirst：当频道存在多个URL地址时，是否优先使用组播地址。可选值：`true`或`false`。**非必填，缺省为`true`**。

* udpxy：当通过启动参数`-u`或`--udpxy`配置了包含内外网的多个udpxy的URL地址时，可通过该参数指定当前m3u所使用的地址。
  **非必填，缺省为其中任意一个URL地址**<br/>

  > 例如，若启动参数配置为：<br/>
  > `./iptv serve -u inner=http://192.168.1.1:4022,outer=http://udpxy.iptv.com:4022`
  > * `/channel/m3u?udpxy=inner`则使用udpxy的内网地址。
  > * `/channel/m3u?udpxy=outer`则使用udpxy的外网地址。
  > * `/channel/m3u?udpxy=notexist`若指定的名称不存在，则使用频道的原始地址。

### txt格式直播源

```
http://IP:PORT/channel/txt?multiFirst={multiFirst}&udpxy={udpxy}
```

#### 参数说明

* multiFirst：参数说明同上。
* udpxy：参数说明同上。

### pls格式直播源

```
http://IP:PORT/channel/pls?multiFirst={multiFirst}&udpxy={udpxy}
```

#### 参数说明

* multiFirst：参数说明同上。
* udpxy：参数说明同上。

### json格式EPG

```
http://IP:PORT/epg/json?ch={name}&date={date}
```  

### xmltv格式EPG

```
http://IP:PORT/epg/xml?backDay={backDay}
```  

#### 参数说明

* backDay：可选保留最近多少天的节目单，**非必填，缺省为查全部**。

### xmltv格式EPG（gzip压缩）

```
http://IP:PORT/epg/xml.gz?backDay={backDay}
```  

#### 参数说明

* backDay：参数说明同上。

## 帮助

* [在OpenWrt中设置自启动](./docs/autostart.md)

# 免责声明

在使用本项目之前，请仔细阅读以下免责声明：

* 本项目的初衷是为研究、学习和技术交流提供帮助，未对其作任何特殊用途的适配。您在使用本项目时，必须遵守适用的法律法规和道德规范。
* 本项目不得用于任何违法或不正当的目的，包括但不限于商业用途、侵权行为或破坏性操作。
* 使用本项目产生的任何后果，由使用者自行承担全部风险和责任。开发者对因使用本项目引发的任何直接或间接损失，不承担任何责任。
* 本免责声明的解释权归项目开发者所有。

**注意：如果您无法接受以上条款，请勿使用或分发本项目。**