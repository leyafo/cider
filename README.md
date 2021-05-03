`cider` is a simple GitHub pages static generator. It's fast and easy to use. See example: https://www.leyafo.com

## Install
`go build -o cider *.go`
or download the zip file from release.

## Generate your site
Add your md file into content folder, your file name must start with date. For example:

````bash
touch content/2021-05-21-your-file-name.md
````

 Generating the entire site:
`./cider `

For exhibiting your site on local HTTP Server
`./cider s`

If you don't want to publish your article, move the md file into `draft` folder. Then running:
```bash
./cider d
```
It will show you article in a local HTTP server. All html files are rendered in `public` folder.

### Publish your site
Create a repository `your_name.github.io` in GitHub. Initialize the public folder
```
git init
git commit -m "first commit"
git remote add origin git@github.com:your_name/your_name.github.io
git push -u origin master
```

### Classify your articles
If you want to classify your articles into different groups, add a subfolder into content, then add the folder link into `templates/partials/_nav.html.tpl`.
For example, if have written an article about English learning, and want to create a English Learning group in your website.
```bash
mkdir content/el
mv your_md_file content/el/
```

Add `el` into `templates/partials/_nav.html.tpl`
```html
[ <a class="nav-btn" href="/el/">English</a> ]
```

### Insert an image into your article
Create a `images` folder in public, then copy the relative path in your article.
```markdown
![your_picture](/images/your_picture.jpg)
```



## License

MIT