package render

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

func Render(tmplPath, contentPath, metaPath, outputPath string){
	renderList := make(RenderList)
	err := readContentInfo(renderList, metaPath)
	if err != nil{
		log.Println(err.Error())
	}
	renderList.UpdateRenderList(contentPath)

	//check the output path does need to remove
	removedPosts := renderList.GetRemovedContentInfo(contentPath)
	for _, p := range(removedPosts){
		removePath := filepath.Join(outputPath, p.IndexKey)
		log.Println("remove path:", removePath)
		if err = os.RemoveAll(removePath); err != nil{
			log.Println(err)
		}
	}

	doesNeedRenderAll := renderList.GetTemplateModifyTimes(tmplPath)
	if doesNeedRenderAll {
		log.Println("Render all files")
	}
	postTemplate := GetTemplate(tmplPath, "post.html.tpl")
	for key, info := range renderList{
		if info.IsContent() && (doesNeedRenderAll || info.NeedRender){
			log.Printf("Rendering %s \n", info.GetMDOutPath(outputPath))
			err := GeneratePost(postTemplate,
				info.GetMDPath(contentPath),
				info.GetMDOutPath(outputPath))
			if err != nil{
				log.Fatal(err.Error())
				break
			}
			renderList[key].NeedRender = false
		}
	}
	log.Println("All md files has been rendered.")

	//generate homepage list, which sort by created time
	log.Println("Render index")
	sepratedList := make(map[string]RenderList)
	for k, v := range renderList{
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
		indexTemplate := GetTemplate(tmplPath, "index.html.tpl")
		err := GenerateList(indexTemplate, v, filepath.Join(outputPath, k, "index.html"))
		if err != nil{
			log.Fatal(err)
		}
	}

	storeContentInfo(renderList, metaPath)
}
