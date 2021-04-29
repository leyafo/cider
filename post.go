package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark-highlighting"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"
)

var(
	markdown goldmark.Markdown
)

func init(){
	markdown = goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithExtensions(
			highlighting.NewHighlighting(
				highlighting.WithStyle("github"),
			),
		),
		goldmark.WithRendererOptions(
			//html.WithHardWraps(),
			html.WithXHTML(),
			html.WithUnsafe(),
		),
	)
}

type Post struct{
	Title string
	Content string
}

func GeneratePost(tmpl *template.Template, mdPath, outPutPath string)error{
	if _, err := os.Stat(outPutPath); os.IsNotExist(err) {
		//make the dir first
		parentDir := filepath.Dir(outPutPath)
		err = os.MkdirAll(parentDir, 0766)
		if err != nil{
			log.Fatal(err.Error())
		}
	}

	outFile, err := os.OpenFile(outPutPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0766)
	if err != nil{
		return err
	}
	return GeneratePostOut(tmpl, mdPath, outFile)
}

func GeneratePostOut(tmpl *template.Template, mdPath string, writer io.Writer)error{
	mdFile, err := ioutil.ReadFile(mdPath)
	if err != nil{
		return err
	}
	title, err := GetTitleFromPostMD(mdPath)
	if err != nil{
		return nil
	}

	var buf bytes.Buffer
	if err = markdown.Convert(mdFile, &buf); err != nil{
		return err
	}
	var p Post
	p.Title = title
	p.Content = buf.String()
	return tmpl.Execute(writer, &p)
}

func GetTitleFromPostMD(mdpath string)(string, error){
	file, err := os.Open(mdpath)
	if err != nil{
		return "", err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan(){
		line := strings.TrimSpace(scanner.Text())
		if len(line) > 0 && line[0] == '#'{
			line = strings.Trim(line, "#")
			return strings.TrimSpace(line), nil
		}
	}
	fmt.Printf("Warning!!! File %s has no title, please add title first.\n", mdpath);

	return "No Title", nil
}

type PostTitleList struct{
	Title string
	Link string
	CreateDate time.Time
	CreateDateStr string
}

type ByDate []PostTitleList

func (a ByDate) Len() int           { return len(a) }
func (a ByDate) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByDate) Less(i, j int) bool {
	if a[i].CreateDateStr == a[j].CreateDateStr{
		return strings.Compare(a[i].Title, a[j].Title) == 1
	}
	return a[i].CreateDateStr > a[j].CreateDateStr
}


func GenerateList(tmpl *template.Template, contentList RenderList, outputFile string)error{
	outFile, err := os.OpenFile(outputFile, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0766)
	if err != nil{
		return err
	}
	return GenerateListOut(tmpl, contentList, outFile)
}

func GenerateListOut(tmpl *template.Template, contentList RenderList, output io.Writer)error{
	var l ByDate
	for _, content := range contentList{
		if content.IsContent(){
			l = append(l, PostTitleList{
				Title:     content.Title,
				CreateDate: content.CreateDate,
				CreateDateStr: content.CreateDate.Format("2006-01-02"),
				Link:  content.IndexKey,
			})
		}
	}
	sort.Sort(&l)
	return tmpl.Execute(output, &l)
}
