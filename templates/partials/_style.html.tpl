{{define "_style"}}
<style>
        body {
            color: #000000;
            line-height: 1.6em;
            padding: 1em;
            margin: auto;
            max-width: 72em;
            background: #fefefe;
            font-family: Fira Code,Monaco,Consolas,Ubuntu Mono,PingFang SC,Hiragino Sans GB,Microsoft YaHei,WenQuanYi Micro Hei,monospace,sans-serif;
        }
        blockquote {
            background: #f9f9f9;
            border-left: 10px solid #ccc;
            margin: 1.5em 10px;
            padding: 0.5em 10px;
            quotes: "\201C""\201D""\2018""\2019";
        }
        blockquote:before {
            color: #ccc;
            content: open-quote;
            font-size: 4em;
            line-height: 0.1em;
            margin-right: 0.25em;
            vertical-align: -0.4em;
        }
        blockquote p {
            display: inline;
        }
        em{
            color: grey;
        }

        pre{
            background-color: #eed;
        }
        code{
            background-color: #eed;
        }
        nav {
            margin-bottom: 20px;
        }
        small{
            font-size: smaller;
        }
        a {
            color: #007bff;
        }

        a:hover,
        a:focus {
            color: #0056b3;
        }

        a:visited{
            color: grey;
        }

        article {
            padding-bottom: 1em;
        }

        img {
            max-width: 100%;
        }

        body {
            background-color: #fff;
            color: #212529;
        }

        .comments{
            text-align: center;
            margin-top: 20px;
            padding: 5px 10px;
            font-size: .8rem;
            text-transform: uppercase;
            text-decoration: none;
            letter-spacing: .1em;
            z-index: 1;
        }

        .utterances{
            max-width: 100%;
        }
        .hr-middle-text{
            line-height: 1em;
            position: relative;
            outline: 0;
            border: 0;
            color: black;
            text-align: center;
            height: 1.5em;
            opacity: .5;
        }
        .hr-middle-text:before {
            content: '';
            background: linear-gradient(to right, transparent, black, transparent);
            position: absolute;
            left: 0;
            top: 50%;
            width: 100%;
            height: 1px;
        }
        .hr-middle-text:after{
            content: attr(data-content);
            position: relative;
            display: inline-block;
            color: black;
            padding: 0 .5em;
            line-height: 1.5em;
            background-color: #fcfcfa;
        }
</style>
{{end}}