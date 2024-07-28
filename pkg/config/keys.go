// Copyright 2022 The imkuqin-zw Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"regexp"
	"strings"
)

var (
	KeyBase = "yggdrasil"

	KeySingleResolver  = "resolver"
	KeySingleBalancer  = "balancer"
	KeySingleEndpoints = "endpoints"
	KeySingleAddress   = "address"
	KeySingleProtocol  = "protocol"
	KeySingleMetadata  = "metadata"

	KeyClient            = Join(KeyBase, "client")
	KeyClientInstance    = Join(KeyClient, "{%s}")
	KeyClientEndpoints   = Join(KeyClient, "{%s}", KeySingleEndpoints)
	KeyClientNamespace   = Join(KeyClientInstance, "namespace")
	KeyClientProtocolCfg = Join(KeyClientInstance, "protocolConfig", "{%s}")
	KeyClientBalancerCfg = Join(KeyClientInstance, "balancerConfig", "{%s}")
	KeyClientInterceptor = Join(KeyClientInstance, "interceptor")
	KeyClientUnaryInt    = Join(KeyClientInterceptor, "unary")
	KeyClientStreamInt   = Join(KeyClientInterceptor, "stream")
	KeyClientIntCfg      = Join(KeyClientInterceptor, "config", "{%s}")

	KeyServer         = Join(KeyBase, "server")
	KeyServerProtocol = Join(KeyServer, "protocol")

	KeyGovernor = Join(KeyBase, "governor")

	KeyRest                 = Join(KeyBase, "rest")
	KeyRestEnable           = Join(KeyRest, "enable")
	KeyRestMarshaler        = Join(KeyRest, "marshaler")
	KeyRestMarshalerSupport = Join(KeyRestMarshaler, "support")
	KeyRestMarshalerCfg     = Join(KeyRestMarshaler, "config", "{%s}")

	KeyInterceptor     = Join(KeyBase, "interceptor")
	KeyIntUnaryClient  = Join(KeyInterceptor, "unaryClient")
	KeyIntStreamClient = Join(KeyInterceptor, "streamClient")
	KeyIntUnaryServe   = Join(KeyInterceptor, "unaryServer")
	KeyIntStreamServer = Join(KeyInterceptor, "streamServer")
	KeyInterceptorCfg  = Join(KeyInterceptor, "config", "{%s}")

	KeyRemoteProto   = Join(KeyBase, "remote.protocol.{%s}")
	KeyRemoteLgLevel = Join(KeyBase, "logger.Logger.level")

	KeyApplication  = Join(KeyBase, "application")
	KeyAppName      = Join(KeyApplication, "name")
	KeyAppRegion    = Join(KeyApplication, "region")
	KeyAppZone      = Join(KeyApplication, "zone")
	KeyAppCampus    = Join(KeyApplication, "campus")
	KeyAppNamespace = Join(KeyApplication, "namespace")
	KeyAppVersion   = Join(KeyApplication, "version")
	KeyAppMetadata  = Join(KeyApplication, "metadata")

	KeyTracer   = Join(KeyBase, "tracer")
	KeyMeter    = Join(KeyBase, "meter")
	KeyRegistry = Join(KeyBase, "registry")

	KeyLogger        = Join(KeyBase, "logger")
	KeyLoggerLevel   = Join(KeyLogger, "level")
	KeyLoggerWriter  = Join(KeyLogger, "writer")
	KeyLoggerTimeEnc = Join(KeyLogger, "timeEncoder")
	KeyLoggerDurEnc  = Join(KeyLogger, "durationEncoder")
	KeyLoggerStack   = Join(KeyLogger, "printStack")

	KeyStats    = Join(KeyBase, "stats")
	KeyStatsCfg = Join(KeyStats, "config")
)

func Join(s ...string) string {
	return strings.Join(s, keyDelimiter)
}

var regx, _ = regexp.Compile(`{([\w.-]+)}`)

func genPath(key, delimiter string) []string {
	matches := make([]string, 0)
	key = regx.ReplaceAllStringFunc(key, func(s string) string {
		matches = append(matches, s[1:len(s)-1])
		return "{}"
	})
	paths := strings.Split(key, delimiter)
	j := 0
	for i, item := range paths {
		if item == "{}" {
			paths[i] = matches[j]
			j++
		}
	}
	return paths
}
