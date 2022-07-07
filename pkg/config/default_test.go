package config

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/imkuqin-zw/yggdrasil/pkg/config/source/file"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test__pointerConfig(t *testing.T) {
	type Config struct {
		Val *int
	}
	cfg := &Config{}
	err := Scan("conf", cfg)
	assert.Nil(t, err)
	assert.Nil(t, cfg.Val)
	err = Set("conf.val", 1)
	assert.Nil(t, err)
	err = Scan("conf", cfg)
	assert.Nil(t, err)
	require.NotNil(t, cfg.Val)
	assert.Equal(t, 1, *cfg.Val)
}

func Test__Scan(t *testing.T) {
	type Config struct {
		Target string
	}
	err := Set("yggdrasil.client.sample.grpc.target", "192.168.3.52:49613")
	require.Nil(t, err)
	c := &Config{}
	key := "yggdrasil.client.sample.grpc"
	err = Scan(key, c)
	fmt.Println(Get(""))
	fmt.Println(Get("yggdrasil"))
	fmt.Println(Get("yggdrasil.client"))
	fmt.Println(Get(key))
	require.Nil(t, err)
	fmt.Println(c)
}

func Test__GetContainerDelimiterKey(t *testing.T) {
	if err := LoadSource(file.NewSource("./testdata/config.yaml", false)); err != nil {
		t.Fatal(err)
	}
	data := Get("yggdrasil.client.{example.polaris.server}.gd").Bytes([]byte("not found"))
	assert.Equal(t, []byte("not found"), data)
	data = Get("yggdrasil.client.{example.polaris.server}").Bytes()
	assert.Equal(t, []byte(`{"grpc":{"target":"127.0.0.1:30001"}}`), data)
}

func Test_matchKey(t *testing.T) {
	key := "yggdrasil.client.{example.polaris.server}"
	regx, _ := regexp.Compile("{([\\w.]+)}")
	matches := make([]string, 0)
	key = regx.ReplaceAllStringFunc(key, func(s string) string {
		matches = append(matches, s[1:len(s)-1])
		return "{}"
	})
	paths := strings.Split(key, ".")
	j := 0
	for i, item := range paths {
		if item == "{}" {
			paths[i] = matches[j]
			j++
		}
	}
	assert.Equal(t, []string{"yggdrasil", "client", "example.polaris.server"}, paths)
}