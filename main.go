package main

import (
	"flag"
	"fmt"
	"github.com/leyafo/cider/render"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

var(
	outputFolder  = flag.String("o", "public", "")
	contentFolder = "content"
	templatePath  = "templates"
	metaPath      = ".meta"
	draftFolder   = "draft"
	indexTemplate = "index.html.tpl"
	postTemplate  = "post.html.tpl"
	addr          = ":8080"
)

func cwdPath(subPath ...string)string{
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


func printHelp(){
	fmt.Println("./cider -o path\n", "render all MD files(in content folder) to path, default is ./public" )
	fmt.Println("./cider s \n", "render all MD files and then start a HTTP server to exhibit your github pages")
	fmt.Println("./cider d \n", "start a HTTP server and render contents in draft folder." )
}

func main(){
	log.SetFlags(log.Lshortfile)

	if len(os.Args) == 2{
		s := viewingServer{}
		s.PostList = make(render.RenderList)
		s.PostTemplate = render.GetTemplate(cwdPath(templatePath),postTemplate)
		s.IndexTemplate = render.GetTemplate(cwdPath(templatePath),indexTemplate)
		switch os.Args[1] {
		case "s":
			s.ContentDir=cwdPath(contentFolder)
			s.PostList.UpdateRenderList(s.ContentDir)
			log.Println("Starting HTTP file server at http://localhost:8080/")
			log.Fatal(http.ListenAndServe(addr, http.HandlerFunc(s.viewingServer)))
			return
		case "d":
			s.ContentDir=cwdPath(draftFolder)
			s.PostList.UpdateRenderList(s.ContentDir)
			log.Println("Starting HTTP file server at http://localhost:8080/")
			log.Fatal(http.ListenAndServe(addr, http.HandlerFunc(s.viewingServer)))
			return
		case "h":
			printHelp()
			return
		default:
			break
		}
	}
	flag.Parse()
	outputPath, _ := filepath.Abs(*outputFolder)
	render.Render(cwdPath(templatePath), cwdPath(contentFolder), filepath.Join(outputPath, metaPath), outputPath)
}

type viewingServer struct {
	ContentDir string
	PostList      render.RenderList
	PostTemplate  *template.Template;
	IndexTemplate *template.Template
}

func (s *viewingServer) viewingServer(w http.ResponseWriter, r *http.Request)  {
	var err error
	p := strings.TrimSpace(r.URL.Path)
	if p[len(p)-1] == '/'{
		err = render.GenerateListWithPath(s.IndexTemplate, s.PostList, p, w)
	}else if strings.Index(p, "/images") == 0{
		f, err := os.Open(cwdPath(p))
		if err != nil{
			w.WriteHeader(http.StatusNotFound)
			return
		}
		io.Copy(w, f)
	}else{
		if p, ok := s.PostList[p]; ok{
			err = render.GeneratePostOut(s.PostTemplate, p.GetMDPath(s.ContentDir), w)
		}else{
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}
	if err != nil{
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}else{
		w.WriteHeader(http.StatusOK)
	}
}
