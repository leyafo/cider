<!DOCTYPE html>

<html lang="zh-cn">
<head>
    <meta charset="UTF-8">
    <title>{{.Title}}</title>
    {{template "_style"}}
    {{template "_script"}}
</head>
<body>
    {{template "_nav"}}
    <article>
        {{ .Content}}
    </article>

    <div class="comments">
        <hr class="hr-middle-text" data-content="COMMENTS">
        <script src="https://utteranc.es/client.js" repo="leyafo/leyafo.github.io" issue-term="title" label="Commnets" theme="github-light" crossorigin="anonymous" async >
        </script>
    </div>
<script>
    document.querySelectorAll("pre").forEach(function (i) {
        i.removeAttribute("style")
    })
</script>
</body>
</html>
