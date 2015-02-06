<!DOCTYPE html>
<html lang="zh-CN">
<head>
	<meta charset="utf-8">
	<meta http-equiv="X-UA-Compatible" content="IE=edge">
	<meta name="description" content="">
	<meta name="HandheldFriendly" content="True">
	<meta name="viewport" content="width=device-width,initial-scale=1,maximum-scale=1,user-scalable=no">

	<link href="/assets/css/screen.css" rel="stylesheet">

	<title>硬盘在歌唱</title>
</head>

<body>
	<header>
		<h1>硬盘在歌唱</h1>
		<p>我编程三日，两耳不闻人声，只有硬盘在歌唱。</p>
	</header>
	
<div class="center-float"><div class="auto-size">

		{{range .}}
		<div class="category-item">
			<h2>{{.Title}}</h2>
			{{if .Description}} <p class="desc">{{.Description}}</p> {{end}}
			{{range .Posts}}
				<p>
					<a href="/{{.Name}}">{{.Title}}</a>
					{{if .Description}} <p class="desc">{{.Description}}</p> {{end}}
				</p>
			{{end}}
		</div>
		{{end}}

		<div class="clear"></div>

		<div class="about">
			<h2>关于</h2>
			<p>黄梦龙，88年生，游戏程序员，现在在北京一个创业团队开发商业游戏。</p>
			<p>本博客使用Go语言开发，目前主要的主题是游戏服务器技术及Go语言，项目代码托管在<a href="https://github.com/huangml/disksing.com">GitHub</a>上。</p>
			<p>这里没有评论系统，有问题可以通过以下各种方式联系到我，文章的问题可以在GitHub项目中提交Issue或PR。</p>
			<ul>
				<li>邮箱：menglong ⓐ outlook.com</li>
				<li>GitHub: <a href="https://github.com/huangml">@huangml</a></li>
				<li>新浪微博：<a href="http://weibo.com/539523448">@黄梦龙烫烫烫</a></li>
			</ul>
		</div>

		<div class="clear"></div>

		<footer>
			<p>本站文章采用<a rel="license" href="http://creativecommons.org/licenses/by/4.0/">CC BY 4.0</a>进行许可，文中涉及代码采用<a rel="license" href="http://creativecommons.org/publicdomain/zero/1.0/">CC0 1.0 Universal</a>进行许可</p>
			<p><a href="http://blog.disksing.com/">硬盘在歌唱</a> &copy; 2015</p>
		</footer>
</div></div>
	
</body>
</html>