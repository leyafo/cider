The `cider` is a simple tool of building GitHub pages. It's fast and easy to use. See example: https://www.leyafo.com

## Install
Compiling from source code, or else you can download the zip file from release.
```bash
go build -o cider *.go  
```

## Generate your site
Add your `md` file into content folder, to apply the article with create time, your file name must start with date. For example:
```bash
touch content/2021-05-20-your-file-name.md
```

Generating the entire site:
```bash
./cider 
```
All html files are rendered in `public` folder.

For exhibiting your site on local HTTP Server.  
```bash
./cider s
```

If you don't want to publish your article and just want to present it as a draft, move the `md` file into `draft` folder. Running:
```bash
./cider d
```
It will show all drafts(in draft folder) on a local HTTP server. 

### Publish your site
Create a repository `your_name.github.io` in GitHub. Initialize the public folder with `your_name.github.io` repository.
```
git init
git commit -m "first commit"
git remote add origin git@github.com:your_name/your_name.github.io
git push -u origin master
```

### Classify your articles
If you want to classify your articles into different groups, you should create a subfolder into content, and then add the folder link into `templates/partials/_nav.html.tpl`. For example, if you have written an article about English learning, and want to create a English Learning group in your website.
```bash
mkdir content/el
mv your_md_file content/el/
```

Add `el` into `templates/partials/_nav.html.tpl`
```html
[ <a class="nav-btn" href="/el/">English</a> ]
```

### Insert an image into your article
Create a `images` folder in public, copy your picture into the folder, write the relative path in your article.
```markdown
![your_picture](/images/your_picture.jpg)
```

## License

MIT
