package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

var(
	contentFolder = "content"
	outputFolder  = "public"
	templatePath  = "templates"
	metaPath      = ".meta"
	draftFolder   = "draft"
)

func Path(subPath ...string)string{
	// using the function
	workDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	for _, p := range subPath{
		workDir = filepath.Join(workDir, p)
	}
	return workDir
}

func GetTemplate(mainTempFile string)*template.Template{
	templ, err:= template.ParseFiles(Path(templatePath, mainTempFile))
	if err != nil{
		log.Fatal(err)
		return nil
	}
	templ, err = templ.ParseGlob(Path(templatePath, "partials", "*"))
	if err != nil{
		log.Fatal(err)
		return nil
	}
	return templ
}

func RenderPosts(list RenderList, isNeedRenderAll bool){
	postTemplate := GetTemplate("post.html.tpl")
	for key, info := range list{
		if info.IsContent() && (isNeedRenderAll || info.NeedRender){
			log.Printf("Rendering %s \n",key)
			err := GeneratePost(postTemplate,
				info.GetMDPath(Path(contentFolder)),
				info.GetMDOutPath(Path(outputFolder)),
			)
			if err != nil{
				log.Fatal(err.Error())
				break
			}
			list[key].NeedRender = false
		}
	}
}

func renderIndexPages(list RenderList){
	sepratedList := make(map[string]RenderList)
	for k, v := range list{
		if !v.IsContent(){
			continue
		}
		paths := strings.Split(k, "/")
		var indexKey string
		if len(paths) == 3{
			indexKey = paths[1]
		}else{
			indexKey = "/"
		}
		if len(paths) > 3{
			panic("unsupported sub dictionary")
		}
		if _, ok := sepratedList[indexKey]; !ok{
			sepratedList[indexKey] = make(RenderList)
		}
		sepratedList[indexKey][k] = v
	}
	for k, v := range sepratedList {
		indexTemplate := GetTemplate("index.html.tpl")
		err := GenerateList(indexTemplate, v, Path(outputFolder, k, "index.html"))
		if err != nil{
			log.Fatal(err)
		}
	}
}


func renderPost(needView bool){
	renderList := make(RenderList)
	err := ReadContentInfo(renderList, metaPath)
	if err != nil{
		renderList.UpdateRenderList(contentFolder)
	}
	renderList.UpdateRenderList(Path(contentFolder))

	//check the output path does need to remove
	removedPosts := renderList.GetRemovedContentInfo(Path(contentFolder))
	for _, p := range(removedPosts){
		removePath := Path(outputFolder, p.IndexKey)
		log.Println("remove path:", removePath)
		if err = os.RemoveAll(removePath); err != nil{
			log.Println(err)
		}
	}

	doesNeedRenderAll := renderList.GetTemplateModifyTimes(Path(templatePath))
	if doesNeedRenderAll {
		log.Println("Render all files")
	}
	RenderPosts(renderList, doesNeedRenderAll)
	log.Println("All md files has been rendered.")

	//generate homepage list, which sort by created time
	log.Println("Render index")
	renderIndexPages(renderList)

	StoreContentInfo(renderList, metaPath)
	if needView{
		log.Println("Starting HTTP file server at http://localhost:8080/")
		s := newViewingServer(contentFolder, outputFolder)
		log.Fatal(http.ListenAndServe(":8080", http.HandlerFunc(s.viewingServer)))
	}
}

func printHelp(){
	fmt.Println("./cider\n", "render all MD files(in content folder) to public" )
	fmt.Println("./cider s \n", "render all MD files and then start a HTTP server to exhibit your github pages")
	fmt.Println("./cider d \n", "start a HTTP server and render contents in draft folder." )
}

func main(){
	log.SetFlags(log.Lshortfile)

	if len(os.Args) == 2{
		switch(os.Args[1]){
		case "s":
			renderPost(true)
		case "d":
			log.Println("Starting HTTP file server at http://localhost:8080/")
			s := newDraftServer(draftFolder)
			log.Fatal(http.ListenAndServe(":8080", http.HandlerFunc(s.draftServer)))
		case "h":
			printHelp()
		default:
			break
		}
	}else{
		renderPost(false)
	}
}

