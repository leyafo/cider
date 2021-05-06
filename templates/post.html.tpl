<!DOCTYPE html>

<html lang="zh-cn">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    {{template "_style"}}
    {{template "_script"}}
</head>
<body>
    {{template "_nav"}}
    <article>
        {{ .Content}}
    </article>

    {{template "_comments"}}
<script>
    document.querySelectorAll("pre").forEach(function (i) {
        i.removeAttribute("style")
    })
</script>
</body>
</html>
