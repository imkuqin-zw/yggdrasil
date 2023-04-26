package governor

import (
	"encoding/json"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/imkuqin-zw/yggdrasil/pkg/config"
)

type ErrResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type setConfigReq struct {
	Keys []string      `json:"keys"`
	Data []interface{} `json:"data"`
}

func respErr(w http.ResponseWriter, code int, err error) {
	w.WriteHeader(code)
	data, _ := json.Marshal(&ErrResponse{
		Code: code,
		Msg:  err.Error(),
	})
	_, _ = w.Write(data)
}

func respSuccess(w http.ResponseWriter, r *http.Request, data interface{}) {
	encoder := json.NewEncoder(w)
	if r.URL.Query().Get("pretty") == "true" {
		encoder.SetIndent("", "    ")
	}
	w.WriteHeader(http.StatusOK)
	if data != nil {
		_ = encoder.Encode(data)
	}
}

func respNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func setConfig(w http.ResponseWriter, r *http.Request) {
	cfg := &setConfigReq{}
	if err := json.NewDecoder(r.Body).Decode(cfg); err != nil {
		respErr(w, http.StatusBadRequest, err)
		return
	}
	if err := config.SetMulti(cfg.Keys, cfg.Data); err != nil {
		respErr(w, http.StatusBadRequest, err)
		return
	}
	respNoContent(w)
	return
}

func getConfig(w http.ResponseWriter, r *http.Request) {
	respSuccess(w, r, json.RawMessage(config.Bytes()))
	return
}

func configHandle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getConfig(w, r)
	case http.MethodPut, http.MethodPost:
		setConfig(w, r)
	}
}

func envHandle(w http.ResponseWriter, r *http.Request) {
	respSuccess(w, r, os.Environ())
}

func newBuildInfoHandle(info *debug.BuildInfo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		respSuccess(w, r, info)
	}
}
