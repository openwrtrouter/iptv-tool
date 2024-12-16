## iptv-tool

IPTV工具，功能列表如下：

* 自动更新频道列表和EPG信息。
* 提供m3u和txt格式直播源在线接口。
* 提供EPG在线接口，支持xmltv和json两种格式。

### 配置说明

```json
{
  "key": "",
  "interfaceName": "",
  "serverHost": "182.138.3.142:8082",
  "ip": "",
  "userID": "",
  "stbType": "",
  "stbVersion": "",
  "stbID": "",
  "areaId": "",
  "mac": "",
  "softwareVersion": "",
  "x-requested-with": ""
}
```

| 字段            | 说明                                                                                                 |
|---------------|----------------------------------------------------------------------------------------------------|
| key           | 不是Authenticator，而是生成Authenticator的秘钥，每个IPTV机顶盒可能都不同，可通过工具根据某次抓包获取的Authenticator反向破解key，具体见下面的使用介绍。 |
| interfaceName | 设备的网络接口名称，和ip字段二选一，优先使用该字段的值。当工具运行在软路由上时，可通过配置自动获取指定接口的IPv4地址。用于获取软路由上某接口被自动分配的IPTV线路的IP地址。        |
| serverHost    | 工具请求的IPTV服务器地址，注意需要走IPTV专用网络才能访问通。                                                                 |
| ip            | 客户端的ip，可任意配置，生成Authenticator所需。当interfaceName已配置时，优先通过interfaceName获取。                             |
| 其他字段          | 均可通过抓包获取，注意x-requested-with可通过抓包HTTP请求头拿到。                                                         |

### 使用介绍

将config.json配置文件与工具放在一起，然后运行工具即可，具体运行命令如下：

* 根据某次抓包获取的Authenticator反向破解key（不依赖配置文件config.json）

```
./iptv key -a xxxxx
```

说明：-a后面指定Authenticator，待运行完毕后会在当前目录下生成key.txt文件，其中可能找到很多key，任意一个均可使用(文件中Find Key后面的即是)。
具体命令参数说明可通过命令查看：

```
./iptv key -h
```

* 直接生成m3u直播源文件（需要配置文件config.json）

```
./iptv channel -f m3u -u http://192.168.3.1:4022
```

说明：运行完毕后会在当前目录下生成iptv.m3u文件，通过-u参数指定软路由的udpxy的http地址
更多命令参数说明可通过命令查看

```
./iptv channel -h
```

* 启动HTTP服务，提供在线m3u和epg接口（需要配置文件config.json）：

```
./iptv serve -i 24h -p 8088 -u http://192.168.3.1:4022
```

说明：-i指定频道和EPG更新间隔时间，-p指定启动的http服务的端口，-u指定udpxy的http地址
更多命令参数说明可通过命令查看

```
./iptv serve -h
```

### HTTP API

* m3u格式直播源在线接口

```
http://IP:PORT/channel/m3u?csFormat={format}
```

参数csFormat可指定回看catchup-source的请求格式，非必填。可选值如下：

| 值 | 是否缺省 | 说明                                                    |
|---|------|-------------------------------------------------------|
| 0 | 是    | `?playseek=${(b)yyyyMMddHHmmss}-${(e)yyyyMMddHHmmss}` |
| 1 | 否    | `?playseek={utc:YmdHMS}-{utcend:YmdHMS}`              |

* txt格式直播源在线接口

```
http://IP:PORT/channel/txt
```  

* json格式EPG

```
http://IP:PORT/epg/json?ch={name}&date={date}
```  

* xmltv格式EPG

```
http://IP:PORT/epg/xml
```  

* xmltv格式EPG（gzip压缩）

```
http://IP:PORT/epg/xml.gz
```  

## 免责声明

在使用本项目之前，请仔细阅读以下免责声明：

* 本项目的初衷是为研究、学习和技术交流提供帮助，未对其作任何特殊用途的适配。您在使用本项目时，必须遵守适用的法律法规和道德规范。
* 本项目不得用于任何违法或不正当的目的，包括但不限于商业用途、侵权行为或破坏性操作。
* 使用本项目产生的任何后果，由使用者自行承担全部风险和责任。开发者对因使用本项目引发的任何直接或间接损失，不承担任何责任。
* 本免责声明的解释权归项目开发者所有。

**注意：如果您无法接受以上条款，请勿使用或分发本项目。**