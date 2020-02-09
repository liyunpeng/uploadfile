package main

import (
	"Irisshow/bn"
	"fmt"
	"github.com/kataras/iris"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	upload_path string = "./share/"
)

//表示一个上传的文件
type Upfile struct {
	Url  string //位置，如share/XXX/
	Name string //文件名，如abc.doc
	Date string //文件时间,ModTime转string
}

//表示权限
type Allow struct {
	UP   bool //可上传
	Down bool //可下载
	Del  bool //可删除
}

type Updir []Upfile //表示一个目录当中，所有上传的文件
// 用Len Less Swap使Updir可排序，可用sort.Sort排序
func (d Updir) Len() int { return len(d) }

// Date降序
func (d Updir) Less(i, j int) bool {
	return d[i].Date > d[j].Date
}

// 交换
func (d Updir) Swap(i, j int) { d[i], d[j] = d[j], d[i] }

func main() {
	fmt.Println("OK!请访问  :8080/share")
	//启动一个http 服务器
	app := iris.New()
	//静态文件服务
	app.HandleDir("/share", "./share")
	//注册视图目录
	tmpl := iris.HTML("./views", ".html")
	app.RegisterView(tmpl)
	//主页
	app.Get("/share", func(ctx iris.Context) {
		var a_list []string
		name := ctx.URLParam("name")
		s := strings.Split(name, ".")
		dirname := strings.Split(name, "/")
		fmt.Println(dirname)
		if len(s) >= 2 {
			ctx.Redirect("/share/down?filename=" + name)
		}

		path := "./share/" + name

		for _, i := range bn.GetAllFile(path) {
			if name != "" {
				a_list = append(a_list, name+"/"+i)
			} else {
				a_list = append(a_list, i)
			}
		}
		ctx.ViewData("a_list", a_list)
		ctx.ViewData("name", name)
		ctx.ViewData("a_list1", bn.GetAllFile("./share"))
		ctx.View("main.html")
	})
	//下载
	app.Get("/share/{path:alphabetical}", func(ctx iris.Context) {
		FlagAllowDel := false //允许删除文件标志
		//URL中的路径
		reqPath := ctx.Path()           //如：/share/aaa
		myfolder := "." + reqPath + "/" //如：./share/aaa/
		//获取执行文件路径：
		rootdir, err := filepath.Abs(filepath.Dir(os.Args[0])) //如：e:\goapp\myapp
		createf := rootdir + reqPath + "/"                     //如：e:\goapp\myapp/share/aaa/
		_, err = os.Stat(createf)                              //os.Stat获取文件信息
		//判断文件夹path存在，否则创建之    ,绝对路径
		if os.IsNotExist(err) {
			os.MkdirAll(createf, os.ModePerm)
		}
		//列出目录下的文件
		var upfile Upfile
		fileins := make(Updir, 0)
		files, _ := ioutil.ReadDir(myfolder)
		for _, file := range files {
			if file.IsDir() {
				continue
			} else {
				upfile.Name = file.Name()
				upfile.Url = ctx.Path() + "/" + file.Name()
				upfile.Date = file.ModTime().Format("2006-01-02 15:04:05")
				fileins = append(fileins, upfile)
			}
		}
		//fmt.Println(fileins[0].Name)
		//倒序排序
		sort.Sort(fileins)
		ctx.ViewData("FlagAllowDel", FlagAllowDel)
		ctx.ViewData("Files", fileins)
		// 渲染视图文件: ./v/index.html
		ctx.View("share.html")

	})
	//主页管理，与主页共用模板 .v/share.html
	app.Get("/admin/{path:alphabetical}", func(ctx iris.Context) {
		FlagAllowDel := true //允许删除文件标志
		//列出目录下的文件
		var upfile Upfile
		fileins := make(Updir, 0)
		myfolder := "./share" + ctx.Path()[6:] + "/"
		files, _ := ioutil.ReadDir(myfolder)
		for _, file := range files {
			if file.IsDir() {
				continue
			} else {
				upfile.Name = file.Name()
				upfile.Url = ctx.Path() + "/" + file.Name()
				upfile.Date = file.ModTime().Format("2006-01-02 15:04:05")
				fileins = append(fileins, upfile)
			}
		}
		//fmt.Println(fileins[0].Name)
		//倒序排序
		sort.Sort(fileins)
		ctx.ViewData("FlagAllowDel", FlagAllowDel)
		ctx.ViewData("Files", fileins)
		// 渲染视图文件: ./v/index.html
		ctx.View("share.html")

	})
	//上传, 接收用XMLHttpRequest上传的文件
	app.Post("/share/{path:alphabetical}", func(ctx iris.Context) {
		//获取文件内容
		file, head, err := ctx.FormFile("upfile")

		//可参考Get时的路径判断pathwww是否存在，这里省略了...
		myfolder := "./Irisshow/" + ctx.Path() + "/"
		defer file.Close()
		//创建文件
		fW, err := os.Create(myfolder + head.Filename)
		fmt.Println(myfolder + head.Filename)
		if err != nil {
			fmt.Printf("文件创建失败%s\n", string(err.Error()))
			return
		}
		defer fW.Close()
		_, err = io.Copy(fW, file)
		if err != nil {
			fmt.Println("文件保存失败")
			return
		}
		ctx.JSON(iris.Map{"success": true, "res": head.Filename})

	})
	//下载,未使用
	app.Get("/share/down", func(ctx iris.Context) {
		//无效ctx.Header("Content-Disposition", "attachment;filename=FileName.txt")
		//dirname:=ctx.URLParam("dirname")
		filename := ctx.URLParam("filename")

		fmt.Println(filename)
		path := "./share/" + filename
		err := ctx.SendFile(path, filename)
		if err != nil {
			fmt.Print(err.Error())
		}
	})

	//删除文件
	app.Post("/admin/{dir}", func(ctx iris.Context) {
		path := ctx.PostValue("path") //如 /admin/aaa/111.txt
		myfolder := "./share" + path[6:]
		fmt.Println(myfolder)
		os.Remove(myfolder)
		ctx.JSON(iris.Map{"success": true, "res": "aaaaaaaaaaaa"})

	})

	app.Run(iris.Addr(":8080"))
}
