package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
)

type DraftServer struct{
	PostList RenderList
	IndexTemplate *template.Template
	PostTemplate *template.Template
	DraftFolder string
}

func newDraftServer(draftFolderDir string)*DraftServer{
	d := new(DraftServer)
	d.PostList = make(RenderList)
	d.IndexTemplate = GetTemplate("index.html.tpl")
	d.PostTemplate = GetTemplate("post.html.tpl")
	d.DraftFolder = draftFolderDir

	d.PostList.UpdateRenderList(draftFolderDir)
	log.Println(d.PostList)
	return d
}

func (d *DraftServer)draftServer(w http.ResponseWriter, r *http.Request){
	var err error
	if r.URL.Path == "/"{
		err = GenerateListOut(d.IndexTemplate, d.PostList, w)
	}else if strings.Index(r.URL.Path, "/images") == 0{
		f, err := os.Open(Path("public", r.URL.Path))
		if err != nil{
			w.WriteHeader(http.StatusNotFound)
			return
		}
		io.Copy(w, f)
	}else{
		log.Println(r.URL.Path)
		if p, ok := d.PostList[r.URL.Path]; ok{
			err = GeneratePostOut(d.PostTemplate, p.GetMDPath(Path(d.DraftFolder)), w)
		}else{
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}
	if err != nil{
		w.WriteHeader(http.StatusInternalServerError)
	}else{
		w.WriteHeader(http.StatusOK)
	}
}

type viewingServer struct {
	publicDir string
	contentDir string
	PostTemplate *template.Template;
}

func newViewingServer(contentDir, publicDir string) *viewingServer{
	d := new(viewingServer)
	d.PostTemplate = GetTemplate("post.html.tpl")
	d.contentDir = contentFolder
	d.publicDir = publicDir

	return d
}

func (s *viewingServer) viewingServer(w http.ResponseWriter, r *http.Request)  {
	path := r.URL.Path;
	if strings.Index(path, "/images") == 0 { //read images
		f, err := os.Open(Path("public", r.URL.Path))
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		io.Copy(w, f)
		return
	}
	mdPath := Path(s.contentDir, path+".md")
	var err error
	if _, err = os.Stat(mdPath); os.IsNotExist(err) {
		// mdfile does not exist, it's a directory
		f, err := os.Open(Path(s.publicDir, path, "index.html"))
		if err != nil{
			w.WriteHeader(http.StatusNotFound)
			return
		}
		io.Copy(w, f)
	}else{
		err = GeneratePostOut(s.PostTemplate, mdPath, w)
	}
	if err != nil{
		w.WriteHeader(http.StatusInternalServerError)
	}else{
		w.WriteHeader(http.StatusOK)
	}
}
