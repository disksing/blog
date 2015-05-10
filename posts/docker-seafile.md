title:  使用Docker+Seafile搭建私有云存储
description: 
time: 2015/05/10 17:48
category: others
++++++++
## 缘起
现如今各种云存储服务其实挺多的，国外有经典的DropBox、Google Drive、微软的OneDrive等，国内也有可以免费使用的各种云。

那么为什么想要搭建私有云存储呢？主要是本着“自己的数据自己管理”的原则。

其一是防止数据被窃取。这些云存储服务往往是和自己的某些平台账号绑定在一起的，或者至少是跟自己的某个邮箱绑定在一起的（密码重设），一旦平台账号或邮箱被黑客获取，所有的数据就一览无余了。再加之网络上社工库泛滥，很多人喜欢在各种网络服务上使用相同 的密码，往往是某一个账号失窃，所有数据全部暴露。

其二是防止数据被主动泄漏。Google退出中国事件之后，我们知道运营在国内的产品数据都是对政府公开的，你一定不想把私密照片传到百度云或是360云盘上去。而“棱镜门”之后，我们知道诸如Google等号称“不作恶”的企业，其数据也是对政府公开的，老大哥时刻盯着你……

其三是防止数据遗失。有些人贪图更便宜的价格或是更大的空间选择不知名的云存储服务，说不定哪天就停止服务了，到时候悔之晚矣。另外貌似诸如百度云如果判定你的视频文件有色情内容，会主动将其清除掉。

这么一看自己搭建私有云存储太有必要性了。至少能保证自己的私人数据与其他互联网账号无关，不被搜索引擎索引，不被政府监视。保证服务器运行并做好数据备份就不会丢失。如果仅在家庭或公司内部使用可以部署在内网，安全系数更高。

## Docker和Seafile介绍
> Docker是一个开源的应用容器引擎，让开发者可以打包他们的应用以及依赖包到一个可移植的容器中，然后发布到任何流行的Linux 机器上，也可以实现虚拟化。容器是完全使用沙箱机制，相互之间不会有任何接口（类似iPhone的 app）。几乎没有性能开销,可以很容易地在机器和数据中心中运行。最重要的是,他们不依赖于任何语言、框架或包装系统。

> 摘自[开源中国][1]

Docker能简化我们的云存储搭建过程，还能使其更安全地运行，更方便的维护。

> Seafile是新一代的开源云存储软件。它提供更丰富的文件同步和管理功能，以及更好的数据隐私保护和群组协作功能。Seafile支持 Mac、Linux、Windows三个桌面平台，支持Android和iOS 两个移动平台。

> Seafile是由国内团队开发的国际型项目，目前已有10万左右的用户，以欧洲用户为多。典型的机构用户包括比利时的皇家自然科学博物馆，德国的Wuppertal气候、能源研究所。

> 摘自[开源中国][2]

开源的云存储软件其实不少，我也先后测试了好几款，最后确认Seafile是目前性能最佳、功能较全、安装最方便的。感谢[海文互知团队][3]令人钦佩的工作成果！

## 搭建指南
### 1. Docker环境
Docker运行的基本需求是Linux  x64，内核版本2.6.32-431或更高版本。具体请按照你的系统参考[官方文档][4]。

本例中我使用的Docker版本是1.6.1，`docker version`运行结果如下。
```
root@iZ255y3f595Z:/home# docker version
Client version: 1.6.0
Client API version: 1.18
Go version (client): go1.4.2
Git commit (client): 4749651
OS/Arch (client): linux/amd64
Server version: 1.6.0
Server API version: 1.18
Go version (server): go1.4.2
Git commit (server): 4749651
OS/Arch (server): linux/amd64
```

### 2. 拉取jenserat/seafile镜像
`jenserat/seafile`镜像（[查看详情][5]）包含了Seafile运行的依赖环境和一些方便的脚本，使用`docker pull jenserat/seafile:latest`拉取该镜像的最新版本。下载完成后输入`docker images`命令可以查看下载到的镜像：
```
root@iZ255y3f595Z:/home# docker images
REPOSITORY          TAG                 IMAGE ID            CREATED             VIRTUAL SIZE
jenserat/seafile    latest              8ef4348733ff        5 weeks ago         325.3 MB
```

### 3. 下载Seafile
`jenserat/seafile`镜像中包含了下载Seafile的脚本，可惜其文件托管在Amazon ECS上，国内无法正常访问。

所以我们需要从[官方网站][6]上手动下载，撰写本文时最高版本是[4.1.2][7]。存放进一个目录并将Seafile解压缩，注意保证此目录所在分区有足够空间（其实也不用太在意，之后空间不足后可以很方便的迁移）。本例中我们把Seafile存放在`/home/app/seafile`。
```
root@iZ255y3f595Z:/home/app/seafile# ls -l
total 19836
-rw-r--r-- 1 root root 20308273 May  5 17:17 seafile-server_4.1.2_x86-64.tar.gz
root@iZ255y3f595Z:/home/app/seafile# tar -zxf seafile-server_4.1.2_x86-64.tar.gz
root@iZ255y3f595Z:/home/app/seafile# ls -l
total 19840
drwxrwxr-x 6  500  500     4096 Mar 29 08:35 seafile-server-4.1.2
-rw-r--r-- 1 root root 20308273 May  5 17:17 seafile-server_4.1.2_x86-64.tar.gz
```

### 4. 配置
使用如下命令启动一个Docker容器来配置Seafile，注意将`/home/app/seafile`换成你的目录！

```
docker run -t -i --rm -p 10001:10001 -p 12001:12001 -p 8000:8000 -p 8080:8080 -p 8082:8082 -v /home/app/seafile:/opt/seafile jenserat/seafile -- /bin/bash
```

容器启动后看到如下输出：
```
*** Running /etc/my_init.d/00_regen_ssh_host_keys.sh...
*** Running /etc/rc.local...
*** Booting runit daemon...
*** Runit started as PID 9
*** Running /bin/bash...
root@635064a090b9:/# May 10 08:37:40 635064a090b9 syslog-ng[20]: syslog-ng starting up; version='3.5.3'
```

下面我们在容器中运行`setup-seafile.sh`脚本后按提示进行配置，本例中我们配置为通过域名`sf.disksing.com`访问，各种端口一路回车用默认的就行，因为我们可以更改启动Docker容器时设置端口映射。

```
root@635064a090b9:/# cd /opt/seafile/seafile-server-4.1.2/
root@635064a090b9:/opt/seafile/seafile-server-4.1.2# ./setup-seafile.sh

You are running this script as ROOT. Are you sure to continue?
[yes|no] yes

-----------------------------------------------------------------
This script will guide you to config and setup your seafile server.

Make sure you have read seafile server manual at

        https://github.com/haiwen/seafile/wiki

Note: This script will guide your to setup seafile server using sqlite3,
which may have problems if your disk is on a NFS/CIFS/USB.
In these cases, we sugguest you setup seafile server using MySQL.

Press [ENTER] to continue
-----------------------------------------------------------------


Checking packages needed by seafile ...

Checking python on this machine ...
Find python: python2.7

  Checking python module: setuptools ... Done.
  Checking python module: python-imaging ... Done.
  Checking python module: python-sqlite3 ... Done.

Checking for sqlite3 ...Done.

Checking Done.


What would you like to use as the name of this seafile server?
Your seafile users will be able to see the name in their seafile client.
You can use a-z, A-Z, 0-9, _ and -, and the length should be 3 ~ 15
[server name]: disksing

What is the ip or domain of this server?
For example, www.mycompany.com, or, 192.168.1.101

[This server's ip or domain]: sf.disksing.com

What tcp port do you want to use for ccnet server?
10001 is the recommended port.
[default: 10001 ]

Where would you like to store your seafile data?
Note: Please use a volume with enough free space.
[default: /opt/seafile/seafile-data ]

What tcp port would you like to use for seafile server?
12001 is the recommended port.
[default: 12001 ]

What tcp port do you want to use for seafile fileserver?
8082 is the recommended port.
[default: 8082 ]


This is your config information:

server name:        disksing
server ip/domain:   sf.disksing.com
server port:        10001
seafile data dir:   /opt/seafile/seafile-data
seafile port:       12001
fileserver port:    8082

If you are OK with the configuration, press [ENTER] to continue.

Generating ccnet configuration in /opt/seafile/ccnet...

done
Successly create configuration dir /opt/seafile/ccnet.

Generating seafile configuration in /opt/seafile/seafile-data ...

Done.

-----------------------------------------------------------------
Seahub is the web interface for seafile server.
Now let's setup seahub configuration. Press [ENTER] to continue
-----------------------------------------------------------------


Creating seahub database now, it may take one minute, please wait...


Done.

creating seafile-server-latest symbolic link ... done


-----------------------------------------------------------------
Your seafile server configuration has been completed successfully.
-----------------------------------------------------------------

run seafile server:     ./seafile.sh { start | stop | restart }
run seahub  server:     ./seahub.sh  { start <port> | stop | restart <port> }

-----------------------------------------------------------------
If the server is behind a firewall, remember to open these tcp ports:
-----------------------------------------------------------------

port of ccnet server:         10001
port of seafile server:       12001
port of seafile fileserver:   8082
port of seahub:               8000

When problems occur, refer to

      https://github.com/haiwen/seafile/wiki

for more information.

```

配置完成后启动Seafile的两个服务测试，seafile是文件管理引擎，seahub提供网页访问服务。seahub首次启动时会要求提供管理员邮箱及密码：
```
root@635064a090b9:/opt/seafile/seafile-server-4.1.2# ./seafile.sh start

Starting seafile server, please wait ...
Seafile server started

Done.
root@635064a090b9:/opt/seafile/seafile-server-4.1.2# ./seahub.sh start

Starting seahub at port 8000 ...

----------------------------------------
It's the first time you start the seafile server. Now let's create the admin account
----------------------------------------

What is the email for the admin account?
[ admin email ] admin@disksing.com

What is the password for the admin account?
[ admin password ]

Enter the password again:
[ admin password again ]



----------------------------------------
Successfully created seafile admin
----------------------------------------



Loading ccnet config from /opt/seafile/ccnet
Loading seafile config from /opt/seafile/seafile-data

Seahub is started

Done.

```

现在用浏览器打开`http://<youdomain_or_ip>:8000`，看到登录页面说明配置完成了。输入`exit`退出并关闭当前容器。
```
root@635064a090b9:/# exit
exit
*** /bin/bash exited with status 0.
*** Shutting down runit daemon (PID 12)...
*** Killing all processes...
```

### 5. 启动Seafile
刚才我们已经在容器内正常启动了Seafile，只是如果每次都要手动操作略显麻烦，所幸`jenserat/seafile`提供了自动调用启动脚本的机制，创建容器时定义`autostart=true`即可。

这次我们给容器取一个有意义的名字`seafile`，如果你想让通过80端口访问，将`-p 8000:8000`改为`-p 80:8000`就可以了。

```
docker run -d \
  --name seafile \
  -p 10001:10001 \
  -p 12001:12001 \
  -p 8000:8000 \
  -p 8080:8080 \
  -p 8082:8082 \
  -v /home/app/seafile:/opt/seafile \
  -e autostart=true \
  jenserat/seafile
```

这份指南到这里就结束了，祝玩得开心！

  [1]: http://www.oschina.net/p/docker
  [2]: http://www.oschina.net/p/seafile
  [3]: http://seafile.com/about/
  [4]: https://docs.docker.com/installation/#installation
  [5]: https://registry.hub.docker.com/u/jenserat/seafile/
  [6]: http://www.seafile.com/download/
  [7]: http://download-cn.seafile.com/seafile-server_4.1.2_x86-64.tar.gz