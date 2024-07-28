package genrest

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"testing"
)

func buildPathVars1(path string) string {
	subPattern0 := regexp.MustCompile(`(?i)^{params[0-9]+}$`)
	if subPattern0.MatchString(path) {
		path = fmt.Sprintf(`%sURLParam(r, "%s")`, "chi.", strings.TrimRight(strings.TrimLeft(path, "{"), "}"))
		return path
	}
	pathValPattern1 := regexp.MustCompile(`(?i)/{(params[0-9]+)}/`)
	path = pathValPattern1.ReplaceAllStringFunc(path, func(subStr string) string {
		params := pathPattern.FindStringSubmatch(subStr)
		return fmt.Sprintf(`/"+%sURLParam(r, "%s")+"/`, "chi.", params[1])
	})
	pathValPattern2 := regexp.MustCompile(`(?i)/{(params[0-9]+)}`)
	path = pathValPattern2.ReplaceAllStringFunc(path, func(subStr string) string {
		params := pathPattern.FindStringSubmatch(subStr)
		return fmt.Sprintf(`/"+%sURLParam(r,"%s")+"`, "chi.", params[1])
	})
	path = fmt.Sprintf(`"%s"`, path)
	path = strings.TrimRight(path, `+""`)
	return path
}

func Test_buildPathVars(t *testing.T) {
	//data, _ := json.Marshal(buildPathVars1("/v1/{parent=pools/*}/users"))
	//fmt.Println(string(data))
	//data, _ = json.Marshal(buildPathVars1("/v1/{name=pools/*/users/*}"))
	//fmt.Println(string(data))
	//data, _ = json.Marshal(buildPathVars1("/v1/poos/{pool}/{name=pools/*/users/*}"))
	//fmt.Println(string(data))
	//return
	{
		path, nameVars := buildPathVars("/v1/{pool}/users")
		nameVarsData, _ := json.Marshal(nameVars)
		fmt.Println(path, string(nameVarsData))
		for k, v := range nameVars {
			fmt.Println(k, buildPathVars1(v))
		}
	}
	fmt.Println()
	{
		path, nameVars := buildPathVars("/v1/{parent=pools/*}/users")
		nameVarsData, _ := json.Marshal(nameVars)
		fmt.Println(path, string(nameVarsData))
		for k, v := range nameVars {
			fmt.Println(k, buildPathVars1(v))
		}
	}
	fmt.Println()
	{
		path, nameVars := buildPathVars("/v1/{name=pools/*/users/*}")
		nameVarsData, _ := json.Marshal(nameVars)
		fmt.Println(path, string(nameVarsData))
		for k, v := range nameVars {
			fmt.Println(k, buildPathVars1(v))
		}
	}
	fmt.Println()
	{
		path, nameVars := buildPathVars("/v1/poos/{pool}/{name=pools/*/users/*}")
		nameVarsData, _ := json.Marshal(nameVars)
		fmt.Println(path, string(nameVarsData))
		for k, v := range nameVars {
			fmt.Println(k, buildPathVars1(v))
		}
	}
}
