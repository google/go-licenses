module github.com/google/go-licenses/testdata/modules/vendoredreplace06

go 1.15

require github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da

replace github.com/golang/groupcache => github.com/mailgun/groupcache v1.3.0
