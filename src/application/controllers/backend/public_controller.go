package backend

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"math/rand"
	"mime/multipart"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/kataras/iris/v12"
	"github.com/xiusin/iriscms/src/application/controllers"
	"github.com/xiusin/iriscms/src/application/models"
	"github.com/xiusin/iriscms/src/application/models/tables"
	"github.com/xiusin/iriscms/src/config"

	"github.com/xiusin/iriscms/src/common/helper"

	"github.com/go-xorm/xorm"
	"github.com/kataras/iris/v12/mvc"
	"github.com/kataras/iris/v12/sessions"
	"github.com/xiusin/iriscms/src/common/storage"
)

type PublicController struct {
	Ctx     iris.Context
	Orm     *xorm.Engine
	Session *sessions.Session
}

func (c *PublicController) BeforeActivation(b mvc.BeforeActivation) {
	b.Handle("ANY", "/upload", "Upload")
	b.Handle("ANY", "/fedir-scan", "FeDirScan")
	b.Handle("ANY", "/attachments", "Attachments")
	b.Handle("ANY", "/ueditor", "UEditor")
}

func (c *PublicController) FeDirScan() {
	c.Ctx.JSON(helper.ScanDir(config.AppConfig().View.FeDirname))
}

//上传图片
func (c *PublicController) Upload() {
	isEditor := true
	settingData := c.Ctx.Values().Get(controllers.CacheSetting).(map[string]string)
	mid := c.Session.GetString("mid")
	if mid == "" {
		mid = "public"
	}
	uploadDir := settingData["UPLOAD_DIR"]
	engine := settingData["UPLOAD_ENGINE"]
	var uploader storage.Uploader
	switch engine {
	case "OSS存储":
		uploader = storage.NewOssUploader(settingData)
	default :
		uploader = storage.NewFileUploader(uploadDir)
		uploadDir = uploader.(*storage.FileUploader).BaseDir
	}
	//生成要保存到目录和名称
	uploadDir = fmt.Sprintf("%s/%s/%s", uploadDir, mid, helper.NowDate("Ymd"))
	file, fs, err := c.Ctx.FormFile("filedata")
	if err != nil {
		if fileData := c.Ctx.FormValue("filedata"); fileData == "" {
			glog.Error("上传文件失败", err.Error())
			uploadAjax(c.Ctx, map[string]interface{}{"state": "打开上传临时文件失败 : " + err.Error(), "errcode": "1",}, isEditor)
			return
		} else {
			dist, err := base64.StdEncoding.DecodeString(fileData)
			if err != nil {
				uploadAjax(c.Ctx, map[string]interface{}{"state": "解码base64数据失败 : " + err.Error(), "errcode": "1",}, isEditor)
			}
			//写入新文件
			f, err := ioutil.TempFile("", "tempfile_"+strconv.Itoa(rand.Intn(10000)))
			if err != nil {
				uploadAjax(c.Ctx, map[string]interface{}{"state": "上传失败 : " + err.Error(), "errcode": "1",}, isEditor)
			}
			f.Write(dist)
			f.Close()
			fo, _ := os.Open(f.Name())
			file = multipart.File(fo)
			defer os.Remove(f.Name())
		}
	}
	defer file.Close()
	var fname string
	var size int64
	if fs != nil {
		size = fs.Size
		fname = fs.Filename
	} else {
		fname = helper.GetRandomString(10) + ".png"
	}

	info := strings.Split(fname, ".")
	ext := strings.ToLower(info[len(info)-1])
	canUpload := []string{"jpg", "jpeg", "png"}
	flag := false
	for _, v := range canUpload {
		if v == ext {
			flag = true
		}
	}
	if !flag {
		uploadAjax(c.Ctx, map[string]interface{}{"state": "不支持的文件类型", "errcode": "1",}, isEditor)
		return
	}
	filename := string(helper.Krand(10, 3)) + "." + ext
	storageName := uploadDir + "/" + filename
	path, err := uploader.Upload(storageName, file)
	if err != nil {
		uploadAjax(c.Ctx, map[string]interface{}{"state": "上传失败:" + err.Error(), "errcode": "1"}, isEditor)
		return
	}
	resJson := map[string]interface{}{
		"originalName": fname,     //原始名称
		"name":         filename,  //新文件名称
		"url":          path,      //完整文件名,即从当前配置目录开始的URL
		"size":         size,      //文件大小
		"type":         "." + ext, //文件类型
		"state":        "SUCCESS", //上传状态
		"errmsg":       path,
		"errcode":      "0",
	}
	if id, _ := c.Orm.InsertOne(&tables.IriscmsAttachments{
		Name:       filename,
		Url:        path,
		OriginName: fname,
		Size:       size,
		UploadTime: time.Now(),
		Type:       models.IMG_TYPE,
	}); id > 0 {
		uploadAjax(c.Ctx, resJson, isEditor)
	} else {
		os.Remove(storageName)
		uploadAjax(c.Ctx, map[string]interface{}{"state":   "保存上传失败", "errcode": "1"}, isEditor)
	}

}

////生成验证码
//func (this *PublicController) VerifyCode() {
//	cpt := captcha.New()
//	fontPath := helper.GetRootPath() + "/resources/fonts/comic.ttf"
//	// 设置字体
//	cpt.SetFont(fontPath)
//	// 返回验证码图像对象以及验证码字符串 后期可以对字符串进行对比 判断验证
//	this.Ctx.ContentType("img/png")
//	img, str := cpt.Create(1, captcha.ALL)
//	this.Session.SetFlash("verify_code", str)
//	png.Encode(this.Ctx.ResponseWriter(), img) //发送图片内容到浏览器
//}

func (c *PublicController) UEditor() {
	action := c.Ctx.URLParam("action")
	switch action {
	case "config":
		c.Ctx.Text("%s", `
{
    "imageActionName": "upload", 
    "imageFieldName": "filedata",
    "imageMaxSize": 2048000, 
    "imageAllowFiles": [".png", ".jpg", ".jpeg", ".gif", ".bmp"],
    "imageCompressEnable": true, 
    "imageCompressBorder": 1600, 
    "imageInsertAlign": "none", 
    "imageUrlPrefix": "", 
    "scrawlActionName": "upload", 
    "scrawlFieldName": "filedata", 
    "scrawlMaxSize": 2048000, 
    "scrawlUrlPrefix": "", 
    "scrawlInsertAlign": "none",
    "catcherLocalDomain": ["127.0.0.1", "localhost", "img.baidu.com"],
    "catcherActionName": "catchimage", 
    "catcherFieldName": "source", 
    "catcherPathFormat": "/ueditor/php/upload/image/{yyyy}{mm}{dd}/{time}{rand:6}", 
    "catcherUrlPrefix": "",
    "catcherMaxSize": 2048000,
    "catcherAllowFiles": [".png", ".jpg", ".jpeg", ".gif", ".bmp"],
    "imageManagerActionName": "attachments-img", 
    "imageManagerUrlPrefix": "",
    "imageManagerInsertAlign": "none", 
    "imageManagerAllowFiles": [".png", ".jpg", ".jpeg", ".gif", ".bmp"],
    "fileManagerActionName": "attachments-file", 
    "fileManagerListPath": "/ueditor/php/upload/file/", 
    "fileManagerUrlPrefix": "", 
    "fileManagerListSize": 20, 
    "fileManagerAllowFiles": [
        ".png", ".jpg", ".jpeg", ".gif", ".bmp",
        ".flv", ".swf", ".mkv", ".avi", ".rm", ".rmvb", ".mpeg", ".mpg",
        ".ogg", ".ogv", ".mov", ".wmv", ".mp4", ".webm", ".mp3", ".wav", ".mid",
        ".rar", ".zip", ".tar", ".gz", ".7z", ".bz2", ".cab", ".iso",
        ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".pdf", ".txt", ".md", ".xml"
    ] 
}
`)
	case "upload":
		c.Upload()
	case "attachments-img":
		c.Ctx.Request().ParseForm()
		c.Ctx.Request().Form.Add("type", models.IMG_TYPE)
		c.Attachments()
	}
}

func uploadAjax(ctx iris.Context, uploadData map[string]interface{}, isEditor bool) {
	ctx.JSON(uploadData)
}

// 读取资源列表
func (c *PublicController) Attachments() {
	page, _ := c.Ctx.URLParamInt64("page")
	if page < 1 {
		page = 1
	}
	start, _ := c.Ctx.URLParamInt64("start")
	if start < 0 {
		start = 0
	}
	var data []*tables.IriscmsAttachments
	attachmentType := c.Ctx.URLParamDefault("type", models.IMG_TYPE)
	cnt, _ := c.Orm.Limit(30, int(start)).Where("`type` = ?", attachmentType).FindAndCount(&data)
	c.Ctx.JSON(map[string]interface{}{
		"state":   "SUCCESS",
		"list":    data,
		"total":   cnt,
		"start":   start,
		"errmsg":  "读取成功",
		"errcode": "0",
	})
}
