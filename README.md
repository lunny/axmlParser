Android binary manifest XML file parser library for Golang
=======

This is a golang port of [github.com/xgouchet/AXML](http://github.com/xgouchet/AXML).

This package aims to parse an android binary mainfest xml file on an apk file.

Installation
------
```
go get github.com/lunny/axmlParser
```

Usage
------

```Go
listener := new(AppNameListener)
_, err := ParseApk(apkfilepath, listener)
```