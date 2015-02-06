title:  Go语言设计模式：组合
description: 
time: 2014/11/17 12:06
category: golang
++++++++

GoF对组合模式的定义是，**将对象组合成树形结构以表示“部分整体”的层次结构，组合模式使得用户对单个对象和组合对象的使用具有一致性**。

对于这句话我是有异议的，这里先卖个关子，我们先从实际例子说起。

组合模式的例子大家都见得很多了，比如文件系统（文件/文件夹）、GUI窗口（Frame/Control）、菜单（菜单/菜单项）等等，我这里也举个菜单的例子，不过不是操作系统里的菜单，是真正的菜单，KFC的……

姑且把KFC里的食物认为是`菜单项`，一份套餐是`菜单`。菜单和菜单项有一些公有属性：名字、描述、价格、都能被购买等，所以正如GoF所说，我们需要一致性地使用它们。它们的层次结构体现在一个菜单里会包含多个菜单项或菜单，其价格是所有子项的和。嗯，这个例子其实不是很恰当，不能很好的体现菜单包含菜单的情况，所以我多定义了一个“超值午餐”菜单，其中包含若干个套餐。

用代码归纳总结一下，最终我们的调用代码是这样的：

```golang
func main() {
	menu1 := NewMenu("培根鸡腿燕麦堡套餐", "供应时间：09:15--22:44")
	menu1.Add(NewMenuItem("主食", "培根鸡腿燕麦堡1个", 11.5))
	menu1.Add(NewMenuItem("小吃", "玉米沙拉1份", 5.0))
	menu1.Add(NewMenuItem("饮料", "九珍果汁饮料1杯", 6.5))

	menu2 := NewMenu("奥尔良烤鸡腿饭套餐", "供应时间：09:15--22:44")
	menu2.Add(NewMenuItem("主食", "新奥尔良烤鸡腿饭1份", 15.0))
	menu2.Add(NewMenuItem("小吃", "新奥尔良烤翅2块", 11.0))
	menu2.Add(NewMenuItem("饮料", "芙蓉荟蔬汤1份", 4.5))

	all := NewMenu("超值午餐", "周一至周五有售")
	all.Add(menu1)
	all.Add(menu2)

	all.Print()
}
```

得到的输出如下：

```
超值午餐, 周一至周五有售, ￥53.50
------------------------
培根鸡腿燕麦堡套餐, 供应时间：09:15--22:44, ￥23.00
------------------------
  主食, ￥11.50
    -- 培根鸡腿燕麦堡1个
  小吃, ￥5.00
    -- 玉米沙拉1份
  饮料, ￥6.50
    -- 九珍果汁饮料1杯

奥尔良烤鸡腿饭套餐, 供应时间：09:15--22:44, ￥30.50
------------------------
  主食, ￥15.00
    -- 新奥尔良烤鸡腿饭1份
  小吃, ￥11.00
    -- 新奥尔良烤翅2块
  饮料, ￥4.50
    -- 芙蓉荟蔬汤1份
```

## 面向对象实现

*先说明一下：Go语言不是面向对象语言，实际上只有struct而没有类或对象。但是为了说明方便，后面我会使用`类`这个术语来表示struct的定义，用`对象`这个术语来表示struct实例。*

按照惯例，先使用经典的面向对象来分析。首先我们需要定义菜单和菜单项的抽象基类，这样使用者就可以只依赖于接口了，于是实现使用上的一致性。

Go语言中没有继承，所以我们把抽象基类定义为接口，后面会由菜单和菜单项实现具体功能：

```golang
type MenuComponent interface {
	Name() string
	Description() string
	Price() float32
	Print()

	Add(MenuComponent)
	Remove(int)
	Child(int) MenuComponent
}
```

菜单项的实现：

```golang
type MenuItem struct {
	name        string
	description string
	price       float32
}

func NewMenuItem(name, description string, price float32) MenuComponent {
	return &MenuItem{
		name:        name,
		description: description,
		price:       price,
	}
}

func (m *MenuItem) Name() string {
	return m.name
}

func (m *MenuItem) Description() string {
	return m.description
}

func (m *MenuItem) Price() float32 {
	return m.price
}

func (m *MenuItem) Print() {
	fmt.Printf("  %s, ￥%.2f\n", m.name, m.price)
	fmt.Printf("    -- %s\n", m.description)
}

func (m *MenuItem) Add(MenuComponent) {
	panic("not implement")
}

func (m *MenuItem) Remove(int) {
	panic("not implement")
}

func (m *MenuItem) Child(int) MenuComponent {
	panic("not implement")
}
```

有两点请留意一下。

1. NewMenuItem()创建的是MenuItem，但返回的是抽象的接口MenuComponent。（面向对象中的多态）
2. 因为MenuItem是叶节点，无法提供Add() Remove() Child()这三个方法的实现，所以若被调用会panic。

下面是菜单的实现：

```golang
type Menu struct {
	name        string
	description string
	children    []MenuComponent
}

func NewMenu(name, description string) MenuComponent {
	return &Menu{
		name:        name,
		description: description,
	}
}

func (m *Menu) Name() string {
	return m.name
}

func (m *Menu) Description() string {
	return m.description
}

func (m *Menu) Price() (price float32) {
	for _, v := range m.children {
		price += v.Price()
	}
	return
}

func (m *Menu) Print() {
	fmt.Printf("%s, %s, ￥%.2f\n", m.name, m.description, m.Price())
	fmt.Println("------------------------")
	for _, v := range m.children {
		v.Print()
	}
	fmt.Println()
}

func (m *Menu) Add(c MenuComponent) {
	m.children = append(m.children, c)
}

func (m *Menu) Remove(idx int) {
	m.children = append(m.children[:idx], m.children[idx+1:]...)
}

func (m *Menu) Child(idx int) MenuComponent {
	return m.children[idx]
}
```

其中`Price()`统计所有子项的`Price`后加和，`Print()`输出自身的信息后依次输出所有子项的信息。另注意`Remove()`的实现（从slice中删除一项）。

好，现在针对这份实现思考下面3个问题。

1. `MenuItem`和`Menu`中都有name、description这两个属性和方法，重复写两遍明显冗余。如果使用其它任何面向对象语言，这两个属性和方法都应该移到基类中实现。可是Go没有继承，这可真是坑爹。
2. 这里我们真正实现了**用户一致性访问**了吗？显然没有，当使用者拿到一个`MenuComponent`后，依然要知道其类型后才能正确使用，假如不加判断在`MenuItem`使用`Add()`等未实现的方法就会产生panic。类似地，我们大可以把文件夹/文件都抽象成“文件系统节点”，可以读取名字，可以计算占用空间，但是一旦我们想往“文件系统节点”中添加子节点时，还是必须得判断它到底是不是文件夹。
3. 接着第2条继续思考：产生某种**一致性访问**现象的本质原因是什么？一种观点：`Menu`和`MenuItem`某种本质上是（is-a）同一个事物（`MenuComponent`），所以可以对它们一致性访问；另一种观点：`Menu`和`MenuItem`是两个不同的事物，只是恰巧有一些相同的属性，所以可以对它们一致性访问。

## 用组合代替继承

前面说到Go语言没有继承，本来属于基类的name和description不能放到基类中实现。其实只要转换一下思路，这个问题是很容易用组合解决的。如果我们认为`Menu`和`MenuItem`本质上是两个不同的事物，只是恰巧有（has-a）一些相同的属性，那么将相同的属性抽离出来，再分别组合进两者，问题就迎刃而解了。

先看抽离出来的属性：

```golang
type MenuDesc struct {
	name        string
	description string
}

func (m *MenuDesc) Name() string {
	return m.name
}

func (m *MenuDesc) Description() string {
	return m.description
}
```

改写`MenuItem`：

```golang
type MenuItem struct {
	MenuDesc
	price float32
}

func NewMenuItem(name, description string, price float32) MenuComponent {
	return &MenuItem{
		MenuDesc: MenuDesc{
			name:        name,
			description: description,
		},
		price: price,
	}
}

// ... 方法略 ...
```

改写`Menu`:

```golang
type Menu struct {
	MenuDesc
	children []MenuComponent
}

func NewMenu(name, description string) MenuComponent {
	return &Menu{
		MenuDesc: MenuDesc{
			name:        name,
			description: description,
		},
	}
}

// ... 方法略 ...
```

**Go语言中善用组合有助于表达数据结构的意图**。特别是当一个比较复杂的对象同时处理几方面的事情时，将对象拆成独立的几个部分再组合到一起，会非常清晰优雅。例如上面的`MenuItem`就是描述+价格，`Menu`就是描述+子菜单。

其实对于`Menu`，更好的做法是把`children`和`Add()` `Remove()` `Child()`也提取封装后再进行组合，这样`Menu`的功能一目了然。

```golang
type MenuGroup struct {
	children []MenuComponent
}

func (m *Menu) Add(c MenuComponent) {
	m.children = append(m.children, c)
}

func (m *Menu) Remove(idx int) {
	m.children = append(m.children[:idx], m.children[idx+1:]...)
}

func (m *Menu) Child(idx int) MenuComponent {
	return m.children[idx]
}

type Menu struct {
	MenuDesc
	MenuGroup
}

func NewMenu(name, description string) MenuComponent {
	return &Menu{
		MenuDesc: MenuDesc{
			name:        name,
			description: description,
		},
	}
}
```

## Go语言的思维方式

以下是本文的重点。使用Go语言开发项目2个多月，最大的感触就是：学习Go语言一定要转变思维方式，转变成功则其乐无穷，不能及时转变会发现自己处处碰壁。

下面让我们用真正Go的方式来实现KFC菜单。首先请默念三遍：没有继承，没有继承，没有继承；没有基类，没有基类，没有基类；接口只是函数签名的集合，接口只是函数签名的集合，接口只是函数签名的集合；struct不依赖于接口，struct不依赖于接口，struct不依赖于接口。

好了，与之前不同，现在我们不是先定义接口再具体实现，因为struct不依赖于接口，所以我们直接实现具体功能。先是`MenuDesc`和`MenuItem`，注意现在`NewMenuItem`的返回值类型是`*MenuItem`。

```golang	
type MenuDesc struct {
	name        string
	description string
}

func (m *MenuDesc) Name() string {
	return m.name
}

func (m *MenuDesc) Description() string {
	return m.description
}

type MenuItem struct {
	MenuDesc
	price float32
}

func NewMenuItem(name, description string, price float32) *MenuItem {
	return &MenuItem{
		MenuDesc: MenuDesc{
			name:        name,
			description: description,
		},
		price: price,
	}
}

func (m *MenuItem) Price() float32 {
	return m.price
}

func (m *MenuItem) Print() {
	fmt.Printf("  %s, ￥%.2f\n", m.name, m.price)
	fmt.Printf("    -- %s\n", m.description)
}
```

接下来是`MenuGroup`。我们知道`MenuGroup`是菜单/菜单项的集合，其`children`的类型是不确定的，于是我们知道这里需要定义一个接口。又因为`MenuGroup`的逻辑是对`children`进行增、删、读操作，对`children`的属性没有任何约束和要求，所以我们这里暂时把接口定义为空接口`interface{}`。

```golang
type MenuComponent interface {
}

type MenuGroup struct {
	children []MenuComponent
}

func (m *Menu) Add(c MenuComponent) {
	m.children = append(m.children, c)
}

func (m *Menu) Remove(idx int) {
	m.children = append(m.children[:idx], m.children[idx+1:]...)
}

func (m *Menu) Child(idx int) MenuComponent {
	return m.children[idx]
}
```

最后是`Menu`的实现：

```golang
type Menu struct {
	MenuDesc
	MenuGroup
}

func NewMenu(name, description string) *Menu {
	return &Menu{
		MenuDesc: MenuDesc{
			name:        name,
			description: description,
		},
	}
}

func (m *Menu) Price() (price float32) {
	for _, v := range m.children {
		price += v.Price()
	}
	return
}

func (m *Menu) Print() {
	fmt.Printf("%s, %s, ￥%.2f\n", m.name, m.description, m.Price())
	fmt.Println("------------------------")
	for _, v := range m.children {
		v.Print()
	}
	fmt.Println()
}
```

在实现`Menu`的过程中，我们发现`Menu`对其`children`实际上有两个约束：需要有`Price()`方法和`Print()`方法。于是对`MenuComponent`进行修改：

```golang
type MenuComponent interface {
	Price() float32
	Print()
}
```

最后观察`MenuItem`和`Menu`，它们都符合`MenuComponent`的约束，所以二者都可以成为`Menu`的`children`，组合模式大功告成！

## 比较与思考

前后两份代码差异其实很小：

1. 第二份实现的接口简单一些，只有两个函数。
2. New函数返回值的类型不一样。

从思路上看，差异很大却也有些微妙：

1. 第一份实现中接口是模板，是struct的蓝图，其属性来源于事先对系统组件的综合分析归纳；第二份实现中接口是一份约束声明，其属性来源于使用者对被使用者的要求。
2. 第一份实现认为`children`中的`MenuComponent`是一种具体对象，这个对象具有一系列方法可以调用，只是其方法的功能会由于子类覆盖而表现不同；第二份实现则认为`children`中的`MenuComponent`可以是任意无关的对象，唯一的要求是他们“恰巧”实现了接口所指定的约束条件。

注意第一份实现中，`MenuComponent`中有`Add()`、`Remove()`、`Child()`三个方法，但却不一定是可用的，能不能使用由具体对象的类型决定；第二份实现中则不存在这些不安全的方法，因为New函数返回的是具体类型，所以可以调用的方法都是安全的。

另外，从`Menu`中取出某个child，其可用方法只有`Price()`和`Print()`，一样可以完全安全的调用。如果想在`MenuComponent`是`Menu`的情况下往其中添加子项呢？很简单：
```golang
if m, ok := all.Child(1).(*Menu); ok {
	m.Add(NewMenuItem("玩具", "Hello Kitty", 5.0))
}
```

清晰明了，如果某child是一个`Menu`，那么我们可以对其进行`Add()`操作。

更进一步，这里我们对类型的要求其实并没有那么强，并不需要它一定要是`Menu`，只是需要其提供组合`MenuComponent`的功能，所以可以提炼出这样一个接口：
```golang
type Group interface {
	Add(c MenuComponent)
	Remove(idx int)
	Child(idx int) MenuComponent
}
```

前面的添加子项的代码改成这样：
```golang
if m, ok := all.Child(1).(Group); ok {
	m.Add(NewMenuItem("玩具", "Hello Kitty", 5.0))
}
```

再考虑一下“购买”这个操作，面向对象的实现中，购买的类型是`MenuComponent`，所以购买操作同时可以应用于`Menu`和`MenuItem`。如果用Go语言的思维方式来考察，可购买对象的唯一要求是有`Price()`，所以购买操作的参数是这样的接口：
```golang
type Product interface {
Price() float32
}
```

于是购买操作不仅可应用于`Menu`和`MenuItem`，还可用于任何提供了价格的对象。我们可以任意添加产品，不论是玩具还是会员卡或者优惠券，只要有`Price()`方法就可以被购买。

## 总结

最后总结一下我的思考：

1. 在组合模式中，一致性访问是个伪需求。一致性访问不是我们在设计时需要去满足的需求，而是当不同实体具有相同属性时自然产生的效果。上面的例子中，我们创建的是menu和MenuItem两种不同的类型，但由于它们具有相同属性，我们能以相同的方式取价格，取描述，加入menu成为子项。
2. Go语言中的多态不体现在对象创建阶段，而体现在对象使用阶段，合理使用“小接口”能显著减少系统耦合度。

PS. 本文所涉及的三份完整代码，我放在play.golang.org上了：（需翻墙）

- 面向对象实现：[http://play.golang.org/p/2DzGhVYseY](http://play.golang.org/p/2DzGhVYseY)
- 使用组合：[http://play.golang.org/p/KuH2Vu7f9k](http://play.golang.org/p/KuH2Vu7f9k)
- Go语言的思维方式：[http://play.golang.org/p/TGjI3CDHD4](http://play.golang.org/p/TGjI3CDHD4)