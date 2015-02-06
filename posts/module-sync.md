title:  游戏逻辑模块组织及数据同步
description: 
time: 2014/09/28 23:20
category: gamedev
++++++++

这周工作主要分两部分，一是服务器这边的基础模块搭建，二是服务器与客户端通信模式以及数据同步等方案的协定和验证。总体来说进展不错。

服务器采用actor模式来构建，目前暂时把服务器上运行着的actor叫做service，每个service维护一个请求队列、一个goroutine不断取出请求并进行处理、一组负责处理消息的逻辑模块。游戏服务器里的每个玩家都是一个service，不隶属玩家的功能模块也作为service运行（如排行榜、聊天、公会），还有其他全局模块也作为独立的service运行（如玩家id生成器、注册模块、GM模块）。不同service之间使用消息进行交互（可为同步或异步），消息默认使用Google Protocol Buffer打包。

本周服务器这边基础功能做了一个简单的日志包，一个暂时能用但是还不靠谱的actor包，一个rest包用来接收回应从客户端发来的http请求。游戏相关的做了玩家注册、登录、数据库加载与存盘、与客户端的数据加载及同步机制。另外做了一个策划配表读取及代码生成的模块。

目前基础模块能跑通但是还很不完善，有些问题还没有想清楚，之后会整理好思路进行优化和测试，最后会进行分享和开源：）

 ------

下面说正题。

本周游戏逻辑模块及服务器客户端数据同步讨论得比较清晰并有了一份良好实现，在这里做一个分享。

一个游戏根据功能可以划分为多个不同的模块，如金钱、背包、装备、技能、任务、成就等。按照软件工程的思想，我们希望分而治之单独实现不同的模块，再将这些模块组合在一起成为一份完整的游戏。但现实是残酷的，不同模块之间往往有千丝万缕的联系，比如购买背包物品会需要扣金币、打一个副本会完成任务，完成任务又会奖励金币和物品，金币的增加又导致一个成就达成。于是我们虽然在不同的类或不同的文件中来实现各个模块，却免不了模块间的交叉引用和互相调用，最后混杂不堪，任何一点小修改都可以导致牵一发而动全身。

 

为了后面说明方便，我们考虑这样一个小型游戏系统：总共有3个模块，分别是金钱、背包、任务。购买背包物品需要消耗金币，卖出背包物品可得到金币，金币增加到一定数额后会导致某个任务的状态变为完成，完成任务可获得物品和金币。这3个模块的调用关系如图。


![1](/assets/img/module-sync-1.png)
 

首先我们把模块的数据和逻辑分离，借鉴经典的MVC模式，数据部分叫作Model，逻辑部分叫作Controller。如此一来，游戏功能部分就被划分出来了两个不同的层次，Controller处于较高的层次上，可以引用一个或者多个Model。Model层专心处理数据，对上层无感知。每个Model都是完全独立的模块，不引用任何Controller或Model，不依赖于其他任何对象，可以单拿出来进行单元测试。

![2](/assets/img/module-sync-2.png)

对于我们的例子，每个模块提供的接口列举如下：

BagModel：获取物品数量，增加物品，扣除物品

MoneyModel：获取金币数量，增加金币，扣除金币

TaskModel：增加任务，删除任务，标记任务为完成

BagController：购买物品，卖出物品

TaskController：完成任务

购买或卖出物品时，由BagController进行或操作校验，随后调用BagModel和MoneyModel完成数据修改。完成任务时，由TaskController调用各个模块。

 

现在唯一的问题是，既然MoneyModel不引用其他模块，那么在金币增加时如何告知任务模块去完成任务呢？这里我们需要引入一个管理依赖的利器：观察者模式。

具体使用方式是把Model实现为一个Subject，对某个Model的数据变化感兴趣的Controller实现为对应的Observer。我们的例子中，MoneyModel是Subject，在金币数量变化时通知所有已注册的Observer；TaskController是MoneyModel的一个Observer，在初始化时向MoneyModel注册。

![3](/assets/img/module-sync-3.png)

注意图中由MoneyModel指向TaskController的虚线箭头，代表MoneyModel数据变化时会去通知TaskController，用虚线是因为MoneyModel并不依赖于TaskController（只依赖于Observer接口）。同样BagModel也可以提供背包物品变化的Subject，如果新加一个任务是要求某物品的数量达某个值，那么TaskController可向BagModel注册，这样在物品变化时就能得到通知了，图中也画出了这条虚线。

对观察者模式不熟悉的读者朋友可以自行查阅资料， 本文的重点并不是介绍设计模式。这里简单提示一下观察者模式的精髓：当某模块调用其他模块时就产生了依赖，这时可以不直接去调用，而是转而实现一个机制，这个机制就是让其他模块告诉自己他们需要被调用。最后调用的流程没变，变化的是依赖关系。

 

在客户端情况要更复杂一些，实际上加入UI后，我们的模块设计就成经典的MVC，这也是我们为什么把数据模块和逻辑模块分别叫Model和Controller的原因。

![4](/assets/img/module-sync-4.png)

这里只画出了背包模块。这里的System API指与游戏运行平台相关的一些接口，可能是操作系统API、引擎API、图形库API等等。View模块和Model模块地位相当，只处理显示而不管游戏功能，需要显示的数据都是由Controller提供的。对于能输入的View同样采用观察者模式，点击等事件发生时通知其他模块（而不是直接调用），注意图中由BagView指向BagController的虚线箭头。

 

下面介绍数据同步的设计。

首先对于网络游戏，客户端所展示的数据是服务器传送过来的。当玩家操作导致数据发生变化时，最好也由服务器更新给客户端。曾经接手过一个项目，很多操作的结果都是客户端先算出来的，于是各种逻辑都是服务器和客户端各实现一遍，很容易两边的数据就不一致了，很让人头疼。

所以我们的同步思路是当客户端向服务器发起一个请求时，服务器将所有变化的数据同步给客户端，客户端收到服务器的返回后再更新数据，绝不私自改动数据。在这个指导思想下，我们消息包结构是这样的（以物品卖出举例）：

```
message BagItemSellCG {
    optional int32 id = 1;
    opitnoal int32 count = 2;
}

message BagItemSellGC {
    optional int32 result = 1;
    optional Sync sync = 2;
    opitonal BagItemSellCG postback = 3;
}
```

服务器向客户端返回的消息几乎总是包含3个字段。result为操作结果可能是0或者错误码，sync中包含了所有的数据更新，postback将客户端的请求消息原封不动返回去，便于客户端进行界面更新或友好提示。

sync是一个比较复杂的message，包含了所有需要更新的Model的数据。感谢Protocol Buffer的optional选项，大多数情况下我们发送的数据只是其中很小的一部分。

先来看服务器端消息处理和同步的设计。

![5](/assets/img/module-sync-5.png)

如图所示，我们在Model和Controller之上新加了一个Handler接口层。Handler负责解析消息包，调用Controller处理消息包，在必要的时候调用SyncController构建同步数据，最后打包成消息返回给客户端。

每个Model在管理数据的基础上会维护变化数据的集合，对于简单的Model比如MoneyModel就是一个bool脏标记，而BagModel则维护变化物品id的集合。变化数据列表在同步之后清除。

 

客户端的结构是类似的。

![6](/assets/img/module-sync-6.png)

与服务器的区别就在于SyncController是负责调用Model更新数据，每个Model都实现数据更新接口。注意除SyncController之外，其他Controller只能读取Model而不能改变其数据，这样就保证了所有数据一定是从服务器同步的。

最后我想以出售物品为例子完整走一遍流程。从客户端进行操作开始，到请求发到服务器，最后再返回客户端更新数据和界面。完整的图比较复杂，混在一起基本上没法看了，只好删掉了客户端的任务模块……

![7](/assets/img/module-sync-7.png)

1. BagView界面产生一个点击，因为BagController是BagView的观察者，所以BagController能得到点击事件的通知。
2. BagController识别出此点击是要出售物品，于是构建好消息包发往服务器。
3. 服务器识别出消息类型是Sell，于是消息被派发给SellHandler。
4. SellHandler调用BagController执行逻辑。
5. BagController取出BagModel和MoneyModel的数据进行条件检查，如果无法执行操作则生成错误码返回给SellHandler，否则调用Model修改数据，此时BagModel会记录下变化物品的id，MoneyModel会做一个脏标记。
6. MoneyModel数据发生变化，通知自己的观察者（TaskController）。
7. TaskController判断任务完成，调用TaskModel更新数据。TaskModel会记录发生变化的任务。
8. SellHandler对BagController的调用返回后，如果出错则直接返回消息包给客户端。否则调用SyncController收集同步数据。
9. SyncController调用各个模块收集同步数据，各个模块提交同步数据后清除自己维护的标记。
10. SellHandler将操作结果和同步数据打包后发往客户端。
11. 客户端识别出消息类型是Sell，消息被派发给SellHandler。
12. BagHandler将消息处理结果发给BagController。
13. BagController根据消息处理结果，通知BagView进行必要的提示。
14. SellHandler将消息包中的数据同步部分发给SyncController。
15. SyncController将同步数据同步给各个模块。
16. BagModel和MoneyModel的数据发生了变化，通知观察者，即对应的Controller。
17. Controller调用View进行界面更新。