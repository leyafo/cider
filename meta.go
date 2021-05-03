package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ContentInfo struct{
	Title string
	Ext string
	IndexKey string
	ModifyTime time.Time
	CreateDate time.Time
	NeedRender bool
}

func (c ContentInfo)IsContent()bool{
	return c.Ext == ".md"
}

func (c ContentInfo)GetMDPath(rootPath string)string{
	return filepath.Join(rootPath, c.IndexKey+c.Ext)
}

func (c ContentInfo)GetMDOutPath(rootPath string)string {
	return  filepath.Join(rootPath, c.IndexKey, "index.html")
}

type RenderList map[string]*ContentInfo
func ReadContentInfo(r RenderList, path string)(error){
	file, err := os.Open(path)
	if err != nil{
		return  err
	}
	defer file.Close()
	return json.NewDecoder(file).Decode(&r)
}

func StoreContentInfo(r RenderList, path string)error{
	log.Println("Save all post update time")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil{
		panic("open file error: "+err.Error())
		return err
	}
	err = json.NewEncoder(file).Encode(r)
	if err != nil{
		panic("open file error: "+err.Error())
		return err
	}
	return  nil
}

func (r RenderList)GetTemplateModifyTimes(templatePath string)bool{
	needRenderALL := false
	filepath.Walk(templatePath, func(path string, fileInfo os.FileInfo, err error) error {
		if fileInfo.IsDir(){
			return nil
		}
		splits := strings.Split(path, templatePath)
		relativePath := splits[1]
		_, ok := r[relativePath]
		if !ok{
			r[relativePath] = &ContentInfo{
				IndexKey:   relativePath,
				ModifyTime: fileInfo.ModTime(),
			}
			needRenderALL = true
			return nil
		}else if !r[relativePath].ModifyTime.Equal(fileInfo.ModTime()){
			r[relativePath].ModifyTime = fileInfo.ModTime()
			needRenderALL = true
			return nil
		}

		return nil
	})
	return needRenderALL
}


func (r RenderList) UpdateRenderList(contentPath string){
	filepath.Walk(contentPath, func(path string, fileInfo os.FileInfo, err error) error {
		if fileInfo.IsDir(){
			return nil
		}

		splitIndex := strings.Index(path, contentPath)
		if(splitIndex == -1){
			log.Fatalf("split path error, content path=%s%t origin path=%s\n", contentPath, path)
			return nil
		}
		relativePath := path[len(contentPath):]
		fileExt := filepath.Ext(path)
		//remove the ext for key
		relativePath = relativePath[:len(relativePath)-len(fileExt)]

		if contentInfo, ok := r[relativePath]; ok{
			// do not need render
			if contentInfo.ModifyTime.Equal(fileInfo.ModTime()){
				return nil
			}
		}
		if fileExt == ".md" {
			mdTitle, err := GetTitleFromPostMD(path)
			if err != nil {
				log.Fatal(err)
				return err
			}
			fileName := filepath.Base(path)
			dateSplits := strings.Split(fileName, "-")

			//parse create time
			d, err := time.Parse("2006-1-2", strings.Join(dateSplits[0:3], "-"))
			if err != nil {
				log.Fatalf("file %s must have create date, %s", fileName, err)
				return err
			}
			r[relativePath] = &ContentInfo{
				Title:      mdTitle,
				Ext:        fileExt,
				IndexKey:   relativePath,
				ModifyTime: fileInfo.ModTime(),
				CreateDate: d,
				NeedRender: true,
			}
		}
		return nil
	})
}

func (c RenderList)GetRemovedContentInfo(contentDir string)[]*ContentInfo{
	var result []*ContentInfo
	for key, info := range c{
		if !info.IsContent(){
			continue
		}
		if _, err := os.Stat(info.GetMDPath(contentDir)); os.IsNotExist(err) {
			result = append(result, info)
			delete(c, key)
		}
	}
	return result
}
