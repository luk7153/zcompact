# zcompact

This plugin is used to merge all the routes in one file.
modify on zhe goctl-go-compact base


### 1. install zcompact

```
$ GOPROXY=https://goproxy.cn/,direct go install https://github.com/luk7153/zcompact@latest
```

### 2. environment setup

Make sure the installed `zcompact` in your `$PATH`

### 3. Usage

```
$ goctl api plugin -p zcompact -api user.api -dir .
```
