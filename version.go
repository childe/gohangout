package main

import "fmt"

var version = `  ebff01c
  Thu May 9 11:48:06 2019 +0800
`
var dependences = `  name = "github.com/aviddiviner/go-murmur"
  revision = "b9740d71e571c1f4ccb570b9bc7f352329d3e600"
  name = "github.com/bkaradzic/go-lz4"
  revision = "74ddf82598bc4745b965729e9c6a463bedd33049"
  name = "github.com/childe/healer"
  revision = "50dcbe896c9f3426a12c53dafab92d5b114ec0dd"
  name = "github.com/eapache/go-xerial-snappy"
  revision = "bb955e01b9346ac19dc29eb16586c90ded99a98c"
  name = "github.com/golang/glog"
  revision = "23def4e6c14b4da8ac2ed8007337bc5eb5007998"
  name = "github.com/golang/snappy"
  revision = "2a8bb927dd31d8daada140a5d09578521ce5c36a"
  name = "github.com/ipipdotnet/datx-go"
  revision = "0ac818a639c339140ca04aaa6ea8a738d9167d28"
  name = "github.com/json-iterator/go"
  revision = "0ff49de124c6f76f8494e194af75bde0f1a49a29"
  name = "github.com/kshvakov/clickhouse"
  revision = "aeedb7b9d0584f393905447021238b5c6e9c5154"
  name = "github.com/modern-go/concurrent"
  revision = "bacd9c7ef1dd9b15be4a9909b8ac7a4e313eec94"
  name = "github.com/modern-go/reflect2"
  revision = "94122c33edd36123c84d5368cfb2b69df93a0ec8"
  name = "gopkg.in/yaml.v2"
  revision = "5420a8b6744d3b0345ab293f6fcba19c978f1183"
`

func init() {
	printVersion()
}

func printVersion() {
	fmt.Println("version:")
	fmt.Println(version)

	fmt.Println("dependences:")
	fmt.Println(dependences)
}
