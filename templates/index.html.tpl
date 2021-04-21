<!DOCTYPE html>

<html lang="zh-cn">
<head>
    <meta charset="UTF-8">
    <link rel="shortcut icon" href="/favicon.ico" />
    <title>Light & Truth</title>
    {{template "_style"}}
    {{template "_script"}}
</head>

<body>
    {{template "_nav"}}
    <article>
        {{range .}}
        <p>
            <a href="{{.Link}}">{{.Title}}</a>
            <br />
            <small>{{.CreateDateStr}}</small>
        </p>
        {{end}}
    </article>
</body>
</html>
