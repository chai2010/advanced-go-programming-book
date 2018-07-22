## 1.8. 配置开发环境

工欲善其事，必先利其器！Go语言编程对外部的编辑工具要求甚低，但是配置适合自己的开发环境却可以达到事半功倍的效果。本节简单介绍几个作者常用的Go语言编辑器和轻量级集成开发环境。

经过多年的发展完善，目前支持Go语言的开发工具已经很多了。其中LiteIDE是国人visualfc用Qt专门为Go语言开发的跨平台轻量级集成开发环境。在早期的Go语言的核心代码库中也包含了vim/Emacs/Netepad++/Eclipse等工具对Go语言支持的各种插件，目前这些第三方扩展已经从核心库剥离到外部仓库去独立维护。相对完整的IDE或插件列表可以从Go语言的官方wiki页面查看: https://github.com/golang/go/wiki/IDEsAndTextEditorPlugins 。

对于Windows环境，Go语言纯代码编写的话推荐Notepad++工具，如果需要代码自动补全和调试的话推荐使用微软的Visual Studio Code集成开发环境。如果是Mac OS X用户，可以选择免费的TextMate编辑器，它被誉为macOS下的Notepad++。如果是想基于iPad Pro平台做轻办公，可以选择收费的Textastic应用，它可以完美地配合Working Copy的Git工作流程，同时支持WebDAV协议。在Linux环境，Go语言纯代码编写的话推荐Gtihub Atom工具，如果是命令行的老司机用户也可以配置自己的Vim/Emacs开发环境，调试环境依然推荐Visual Studio Code。

## Windows: Notepad++

Notepad++是Windows操作系统下严肃程序员们编写代码的利器！Notepad++不仅仅免费、体积小（安装程序7+MB）、启动迅速，而且对中文的各种编码支持非常友好，支持众多程序语言的语法高亮，对于正则表达式、函数列表、多工程等特性也有不错的支持。

首先去Notepad++的官网 http://notepad-plus-plus.org 下载最新的安装包。然后去 https://github.com/chai2010/notepadplus-go 下载针对Go语言的配置文件并安装。需要说明的是，对于Go汇编语言用户来说，notepadplus-go是目前唯一支持Go汇编语言语法高亮和函数列表的开发环境。

下面是Go语言的语法高亮预览，其中右侧是Go函数列表：

![](../images/ch1-08-npp-go.png)

下面是Go语言汇编的语法高亮预览：其中右侧是汇编函数列表:

![](../images/ch1-08-npp-go-asm.png)

对于Protobuf或GRPC的用户，可以从 https://github.com/chai2010/notepadplus-protobuf 下载相应的插件。

下面是Protobuf的语法高亮预览：

![](../images/ch1-08-npp-proto.png)

**配置Notepad++的语法高亮**

Notepad++从v6.2版本之后，用户自定义语言文件`userDefineLang.xml`改用`UDL2`语法，这些的配置文件全部采用的是新的`UDL2`的语法。

如果是通过Notepad++安装程序安装的，需要将`userDefineLang.xml`文件中的内容添加到`%APPDATA%\Notepad++\userDefineLang.xml`文件中，放在`<NotepadPlus> ... </NotepadPlus>`标签中间，然后重启Notepad++程序。

如果是从Notepad++ zip/7z压缩包解压绿色安装，配置文件`userDefineLang.xml`在解压目录。


**配置Notepad++函数列表支持**

函数列表功能是Notepad++ v6.4新增加的特性，配置方法和语法高亮的配置过程类似。需要注意的是v6.4和v6.5对应的`<associationMap>...</associationMap>`配置语法稍有不同，具体请参考`functionList.xml`文件中的注释说明。

如果是采用Notepad++安装程序安装的，需要将`functionList.xml`文件中的内容添加到`%APPDATA%\Notepad++\functionList.xml`文件中，放到`<associationMap> ... </associationMap>`和`<parsers> ... </parsers>`标签中间，然后重启Notepad++程序。

如果是从Notepad++ zip/7z压缩包解压绿色安装，配置文件`functionList.xml`在解压目录。

**Notepad++内置函数的自动补全**

Notepad++还支持关键字的自动补全。假设Notepad++安装在`<DIR>`目录，将`go.xml`文件复制到`<DIR>\plugins\APIs`目录下，然后重启Notepad++程序。

下面是内置函数`println`自动补全后函数参数提示的预览图:

![](../images/ch1-08-npp-auto-completion.png)

这是一个比较鸡肋的功能，建议用户根据自己需要选择安装。

## 命令行窗口

对于Go语言开发来说，需要经常在命令行运行`go fmt`、`go test`、`go run x.go`等辅助工具。虽然Notepad++也可以将这些工具配置成标准的菜单中，但是命令行依然是不可缺少的开发环境。

不过Windows自带的命令行工具比较简陋，不是理想的命令行开发环境。如果读者还没有自己合适命令行环境，可以试试ConEmu这个免费命令行软件。ConEmu支持多标签页窗口，复制粘贴也比较方便。下面是ConEmu的预览图：

![](../images/ch1-08-ConEmu.png)


ConEmu的主页在：http://conemu.github.io/ 。

## macOS: TextMate

对于macOS平台的用户，免费的轻量级编辑器软件推荐TextMate。TextMate是macOS下的Notepadd++工具。支持目录列表，支持Go语言的诸多特性。下面是TextMate的预览图：

![](../images/ch1-08-TextMate-1.png)

对于iPad Pro用户，目前也有不少编辑软件对Go语言提供了不错的支持。比如Textastic Code、Coda等，很多都支持iPad和macOS平台的同步，它们一般都是需要单独购买的收费软件。


## iOS: Textastic

Textastic是一款收费应用，它是macOS/iOS下著名的轻量级代码编辑工具，支持包含Go语言在内的多达80多种编程语言的高亮显示。Textastic功能特点有：

- 句法高亮，同时支持80余种语言
- 与TextMate句法定义，主题兼容
- 对HTML、CSS、JavaScript、PHP、C、Objective-C支持自动补全代码
- Symbol list快速导航内容
- 自动保存代码和版本
- iCloud同步 (Mountain Lion only)

在macOS下，Textastic的界面和TextMate非常相似。不过Textastic在左边侧栏提供了基于工程的检索工具。下面是macOS下Textastic的预览图：

![](../images/ch1-08-macos-textastic.png)

因为iOS环境不支持编译和调试，如果需要在iOS环境编写Go程序，首先要解决和其他平台的共享问题。这样可以在iOS环境编写代码，然后在其他电脑上进行编译和测试。

最简单的共享方式是在iCloud的Textastic专有的目录中创建Go语言的工作区目录，然后通过iCloud方案实现和其他平台共享。此外，还可以通过WebDAV标准协议来实现文件的共享，常见的NAS系统都会提供WebDAV协议的共享方式。另外，用Go语言也能很容易实现一个WebDAV的服务器，具体请参考第七章中WebDAV的相关主题。

下面是iPad Pro下Textastic的预览图：

![](../images/ch1-08-ios-textastic-02.png)

如果Go语言代码是放在Git服务器中，可以通过Working Copy应用将仓库克隆到iOS中，然后再Textastic中通过iOS协议打开工作区文件。编辑完成之后，在通过Working Copy将修改提交到中心仓库中。

下面是iPad Pro下Working Copy查看Git更新日志的预览图：

![](../images/ch1-08-ios-working-copy-02.png)


## 跨平台编辑器: Github Atom

Gtihub Atom是Github专门为程序员推出的一个跨平台文本编辑器。具有简洁和直观的图形用户界面，内置支持Go语言语法高亮。同时Github Atom支持宏、自动完成分屏功能，同时集成了文件管理器，对于macOS和Linux用户来说是一个优秀的Go语言编辑器。

Github Atom的预览图:

![](../images/ch1-08-atom-01.png)

Github Atom作为一个Go语言编辑器，不足之处是没有Go汇编语言的高亮显示插件。而且Github Atom对于大文件的支持，性能不是很好。

## 跨平台IDE: Visual Studio Code

Visual Studio Code是微软推出的轻量级跨平台集成开发环境，简称VSCode。VSCode最初的目的是支持JavaScript和TypeScript开发，但是它逐步增加了第三方编程语言的支持，目前它已经可以说是最完美的Go语言集成开发环境了。

VSCode虽然是基于Gtihub Atom而来，不过VSCode支持Go语言的代码自动补全和调试功能，因此已经超越Github Atom单作为编辑器的定位，是一个轻量级的集成开发环境。VSCode对于大文件的支持也比Gtihub Atom优秀很多。

下面是用VScode打开的Go语言工程的预览图:

![](../images/ch1-08-vscode-01.png)

因为，VSCode和Gtihub Atom都是采用的Chrome核心，它不仅仅能编辑显示代码，还可以用来显示网页查看图像，甚至可以在一个分屏窗口中播放视频文件：

![](../images/ch1-08-vscode-02.jpg)

VSCode安装Go语言插件中，默认的很多参数设置比较严格。比如，默认会使用`golint`来严格检查代码是否符合编码规范，对于git工程启动时还会自动获取和刷新。对于一般的Go语言代码来说，`golint`检测过于严格，很难完全通过（Go语言标准库也无法完全通过），从而导致每次保存时都会提示很多干扰信息。当然，对相对稳定的程序定期做`golint`检查也是有必要，它的信息可以作为我们改进代码的参考。同样的，如果git仓库有密码认证的话，VSCode在启动的时候总是弹出输入密码的对话框。

我们可以在工程目录的`.vscode/settings.json`配置文件中定制这些选项。下面配置是强制在保存的时候采用`gofmt`格式化代码，并且关闭保存时`golint`检查。同时在VSCode刚启动的时候，禁止Git自动刷新和获取操作。

```json
// 将设置放入此文件中以覆盖默认值和用户设置。
{
	// Pick 'gofmt', 'goimports' or 'goreturns' to run on format.
	"go.formatTool": "gofmt",

	// [EXPERIMENTAL] Run formatting tool on save.
	"go.formatOnSave": true,

	// Run 'golint' on save.
	"go.lintOnSave": false,

	"git.autorefresh": false,
	"git.autofetch": false
}
```

VSCode作为一个专业的Go语言集成开发环境，稍显不足之处是没有Go汇编语言的高亮显示插件。
