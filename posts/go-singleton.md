title:  Go语言设计模式：单例
description: 
time: 2015/06/06 04:48
category: languages
++++++++

老实说，我觉得单例是23种设计模式里最没意思的了，甚至都算不上是个模式。但因为Go语言有些特立独行的package设计，导致Go语言中的单例写起来跟其它语言不太一样，所以还是值得一聊的。再加上这个系列已经很久没更新了，算是水一个吧：）

## 1. 面向对象与全局变量
通常面向对象软件设计方法中对全局变量和全局函数的使用是比较谨慎的，一般认为这二者的使用会破坏对象的封装性，从而容易导致糟糕的设计。

Java作为一个相对“纯”的面向对象语言直接就废弃了全局变量和全局函数，可实际上它们本质上是不可能被完全消除的。

理想的面向对象世界里一切皆是对象，所有对象都经历“被创建->与其他对象发生交互->走向消亡”的过程。可问题是一切皆对象，那一开始由谁来创建他们呢，于是就产生了`class App`里套一个`public static Main()`的诡异现象。

`public static`是什么鬼？不就是全局的嘛，只不过访问时前面要加个`ClassName.`而已。`ClassName`不是个对象，是类型，隶属于类型系统管辖，任意模块只要import进去随便用，和全局的没什么分别。

其实单例模式这么被发明出来的：某些模块从逻辑上来看自成一体，不依赖于系统其它模块，既不能被其它模块包含，也不应被其它模块创建。简而言之，这个模块就应该是个全局变量。可是面向对象中不能再搞全局变量这套了啊！一定要封装啊！高内聚啊！于是用static成员来保存这个本来应该是全局的变量，再用static成员函数来获取/创建。

这就是我为什么说觉得单例都算不上是个设计模式，因为它并不是在解决软件开发的问题，只是在应付编程语言设计的坑。

基于以上分析，显然Go语言在全局变量上的放任态度并不应被认为是倒行逆施，而是没有跟随Java们的脚步在错误的道路上越行越远……

按惯例接下来想个案例上代码，我们实现一个全局的计数器，提供增加计数、获取计数两个接口。Go语言中封装的边界是package，首字符为小写的变量都不会暴露至package外部，本例中我们使用单独的package来实现计数器。注意其中读写锁的使用保证并发环境下计数的准确性。

```golang
package counter

import "sync"

var (
        number int
        mtx    sync.RWMutex
)

func Add(n int) {
        mtx.Lock()
        defer mtx.Unlock()

        number += n
}

func Get() int {
        mtx.RLock()
        defer mtx.RUnlock()

        return number
}
```

## 2. `init()`函数
上文中的简单记数器总是从0开始计数，实践中的单例一般都需要某种初始化的过程，比如计数器的初始值可能是启动时从文件中加载的。

Go语言提供了`init()`函数这个非常方便的初始化设施，`init()`的运行时机在`main()`运行之前，并且是从`main`包开始依照import依赖关系依次递归初始化的。利益于Go语言各个包之间不允许出现循环依赖的设定，各包之间的依赖关系是一个非常干净的树形结构，几乎不会出来初始化顺序导致的各种问题（C++程序员应该都被坑过的……）。

另外，上文的例子中使用一个单独的小包来做计数器，封装性控制得确实很好。但是工程实践如果大量创建这样的小包，很容易使项目变得难于维护，所以更现实的做法是我们把统计相关设施组织成一个statistic包，我们的计数器是其中的一个文件。

示例代码如下。这里大略解释下为什么要定义`counter`这个struct。

其一，我们现在的包是statistic，原先的两个变量`number`和`mtx`在counter包里的意义很明确，但如果作为statistic包内的全局变量就会对包内的其它功能造成干扰，即使不移至struct内，也要改名为`counterNumber`和`counterMtx`。

其二，上例中counter接口的调用方式是`counter.Get()`和`counter.Add()`，看上去特别直观，一旦包名变为statistic后就不像话了。封装进struct后调用方式是`statistic.Counter.Get()`，这样就对劲多了。

```golang
package statistic

import "sync"

type counter struct {
        number int
        mtx    sync.RWMutex
}

func (c *counter) Add(n int) {
        c.mtx.Lock()
        defer c.mtx.Unlock()

        c.number += n
}

func (c *counter) Get() int {
        c.mtx.RLock()
        defer c.mtx.RUnlock()

        return c.number
}

var Counter *counter

func loadNumber() (n int, err error) {
        //...
        return
}

func init() {
        if n, err := loadNumber(); err != nil {
                panic(err)
        } else {
                Counter = &counter{
                        number: n,
                }
        }
}
```

## 3. 惰性初始化（lazy initialize）
惰性初始化也是单例的常见用法，即单例初始化的时机不是程序启动时，而是第一次使用时，有助于改善程序启动时的footprint和速度。

本例使用`sync.once`来确保并发环境下初始化总是执行一次。

```golang
package statistic

import "sync"

type Counter struct {
        number int
        mtx    sync.RWMutex
}

func (c *Counter) Add(n int) {
        c.mtx.Lock()
        defer c.mtx.Unlock()

        c.number += n
}

func (c *Counter) Get() int {
        c.mtx.RLock()
        defer c.mtx.RUnlock()

        return c.number
}

func loadNumber() (n int, err error) {
        //...
        return
}

var (
        counter *Counter
        once    sync.Once
)

func LoadCounter() *Counter {
        once.Do(func() {
                if n, err := loadNumber(); err != nil {
                        // log error or panic
                } else {
                        counter = &Counter{
                                number: n,
                        }
                }
        })

        return counter
}
```
## 4. 一些思考
总体来看，Go语言的单例写出来与其它语言的版本大相径庭，甚至初看不会觉得这就是单例。

我个人认为单例模式本就是不该有的，正常的思维应该是“家里只需要用一个电冰箱，那我就只买一个用好了”，而决不是“家里只需要用一个电冰箱，我发现这个事情是有规律的，除了电冰箱，洗衣机电视机都只需要一个，我应该有个账本，每次买家电的时候先检查一下，买完了立即记录”……

Go语言package的依赖关系处理得很好，但实践中似乎并不是太容易掌握合适的粒度和划分边界。另外调用包内函数的形式`package.Func()`与调用struct上方法的`struct.Func()`看起来一模一样，时常影响代码阅读，写代码时最好仔细选择包名和公有函数名提高调用处代码的可读性。

Go语言的`sync`包简洁好用，多写简单的几行让你的代码并发安全，很划算。