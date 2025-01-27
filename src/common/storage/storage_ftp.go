package storage

import (
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jlaffaye/ftp"
	"github.com/xiusin/pine"

	"github.com/xiusin/pine/di"
	"github.com/xiusin/pinecms/src/application/controllers"
	"github.com/xiusin/pinecms/src/common/helper"
	"github.com/xiusin/pinecms/src/config"
)

type FtpUploader struct {
	sync.Mutex

	host         string
	ftpUrlPrefix string
	client       *ftp.ServerConn
	conn         func() error
}

var _ Uploader = (*FtpUploader)(nil)

func NewFtpUploader(opt map[string]string) *FtpUploader {
	timeout, _ := strconv.Atoi(opt["FTP_CONN_TIMEOUT"])
	if timeout == 0 {
		timeout = 5
	}
	prefix := strings.TrimSuffix(opt["FTP_URL_PREFIX"], "/")

	uploader := &FtpUploader{host: opt["PROXY_SITE_URL"], ftpUrlPrefix: prefix}

	uploader.conn = func() error {
		c, err := ftp.Dial(
			opt["FTP_SERVER_URL"]+":"+opt["FTP_SERVER_PORT"],
			ftp.DialWithTimeout(time.Duration(timeout)*time.Second),
			ftp.DialWithDisabledUTF8(true),
			ftp.DialWithDisabledEPSV(true),
		)
		if err != nil {
			return err
		}
		if err := c.Login(opt["FTP_USER_NAME"], opt["FTP_USER_PWD"]); err != nil {
			c.Quit()
			return err
		}
		uploader.client = c
		return nil
	}

	helper.PanicErr(uploader.conn())

	pine.RegisterOnInterrupt(func() {
		uploader.client.Logout()
		uploader.client.Quit()
	})

	go func() {
		for range time.Tick(time.Second * 5) {
			uploader.Lock()
			if err := uploader.client.NoOp(); err != nil {
				pine.Logger().Warn("ftp链接错误", err.Error())
				uploader.conn()
			}
			uploader.Unlock()
		}
	}()

	return uploader
}

func (s *FtpUploader) GetEngineName() string {
	return "FTP存储"
}

func (s *FtpUploader) Upload(storageName string, LocalFile io.Reader) (string, error) {
	s.Lock()
	defer s.Unlock()
	storageName = getAvailableUrl(storageName)
	if err := s.client.Stor(storageName, LocalFile); err != nil {
		return "", fmt.Errorf("上传文件%s错误: %s", storageName, err.Error())
	}
	return storageName, nil
}

func (s *FtpUploader) List(dir string) ([]File, error) {
	s.Lock()
	entities, err := s.client.List(dir)
	s.Unlock()

	if err != nil {
		return nil, err
	}
	fileArr := []File{}
	for _, entity := range entities {
		if entity.Name == "." || entity.Name == ".." {
			continue
		}
		f := File{
			Id:       getAvailableUrl(filepath.Join(dir, entity.Name)),
			FullPath: getAvailableUrl(filepath.Join(dir, entity.Name)),
			Name:     entity.Name,
			Size:     int64(entity.Size),
			Ctime:    entity.Time,
		}
		if entity.Type.String() == "folder" {
			f.IsDir = true
		}
		fileArr = append(fileArr, f)
	}
	return fileArr, nil
}

func (s *FtpUploader) Exists(name string) (bool, error) {
	s.Lock()
	_, err := s.client.FileSize(getAvailableUrl(name))
	s.Unlock()
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (s *FtpUploader) GetFullUrl(name string) string {
	if len(s.ftpUrlPrefix) > 0 {
		return s.ftpUrlPrefix + "/" + strings.TrimPrefix(getAvailableUrl(name), "/")
	}
	return strings.TrimRight(s.host, "/") + getAvailableUrl(name)
}

func (s *FtpUploader) Remove(name string) error {
	s.Lock()
	defer s.Unlock()
	return s.client.Delete(getAvailableUrl(name))
}

func (s *FtpUploader) Mkdir(dir string) (err error) {
	s.Lock()
	defer s.Unlock()
	dir = getAvailableUrl(dir)
	dirs := strings.Split(strings.Trim(dir, "/"), "/")
	section := make([]string, len(dirs))
	for _, dir := range dirs {
		section = append(section, dir)
		err = s.client.MakeDir(strings.Join(section, "/"))
	}
	return
}

func (s *FtpUploader) Rmdir(dir string) error {
	s.Lock()
	defer s.Unlock()

	return s.client.RemoveDirRecur(getAvailableUrl(dir))
}

func (s *FtpUploader) Content(name string) ([]byte, error) {
	s.Lock()
	defer s.Unlock()
	resp, err := s.client.Retr(getAvailableUrl(name))
	if err != nil {
		return nil, err
	}
	defer resp.Close()

	return ioutil.ReadAll(resp)
}

func (s *FtpUploader) Rename(oldname, newname string) error {
	s.Lock()
	defer s.Unlock()

	return s.client.Rename(getAvailableUrl(oldname), getAvailableUrl(newname))
}

func init() {
	di.Set(fmt.Sprintf(controllers.ServiceUploaderEngine, (&FtpUploader{}).GetEngineName()), func(builder di.AbstractBuilder) (any, error) {
		cfg, err := config.SiteConfig()
		if err != nil {
			return nil, err
		}
		return NewFtpUploader(cfg), nil
	}, false)
}
