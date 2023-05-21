# my_blog

这是一个使用 Go 编写的简单博客网站，它连接了 MySQL 数据库，并使用了 HTTP 路由和模板渲染。

该程序主要有以下功能：

首页：显示博客文章列表

创建文章：提供表单用于创建新的博客文章

用户注册：提供表单用于用户注册

用户登录：提供表单用于用户登录，并设置 Cookie
以下是程序文件的基本说明：

main.go：包含程序的入口点函数及所有的路由处理函数。

/templates 文件夹：包含所有使用的 HTML 模板。

/static 文件夹：包含静态资源文件，例如 CSS 样式表和 JavaScript 脚本等。

go.mod 和 go.sum：Go 语言的依赖管理文件。