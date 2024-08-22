package rest

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
	"github.com/imkuqin-zw/yggdrasil/pkg/metadata"
	"github.com/imkuqin-zw/yggdrasil/pkg/metadata/peer"
	"github.com/imkuqin-zw/yggdrasil/pkg/rest/marshaler"
	"github.com/imkuqin-zw/yggdrasil/pkg/rest/middleware"
	"github.com/imkuqin-zw/yggdrasil/pkg/status"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xarray"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xnet"
	"github.com/pkg/errors"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/protobuf/proto"
)

type Config struct {
	Host         string
	Port         int
	AcceptHeader string
	OutHeader    string
	OutTrailer   string
	Middleware   struct {
		Rpc string
		Web string
		All string
	}
}

type serverInfo struct {
	address    string
	attributes map[string]string
}

func (s *serverInfo) GetAttributes() map[string]string {
	return s.attributes
}

func (s *serverInfo) GetAddress() string {
	return s.address
}

// ServeMux is a request multiplexer for RPC-gateway.
// It matches http requests to patterns and invokes the corresponding handler.
type ServeMux struct {
	chi.Router
	rpcRouter chi.Router
	webRouter chi.Router
	svr       *http.Server
	mu        sync.Mutex
	listener  net.Listener
	stopped   bool
	started   bool

	info *serverInfo

	acceptHeaders []string
	outHeaders    []string
	outTrailers   []string
}

func NewServer() Server {
	cfg := &Config{}
	if err := config.Get(config.KeyRest).Scan(cfg); err != nil {
		logger.FatalField("fault to load rest config", logger.Err(err))
	}

	ip, _ := xnet.Extract(cfg.Host)
	address := fmt.Sprintf("%s:%d", ip, cfg.Port)

	r := chi.NewMux()

	allMiddlewares := xarray.RemoveDuplicates(
		strings.Split(cfg.Middleware.All, ","),
	)
	r.Use(middleware.GetMiddlewares(allMiddlewares...)...)

	rpcMiddlewares := xarray.RemoveDuplicates(strings.Split(
		"marshaler,"+cfg.Middleware.Rpc, ",",
	))

	webMiddlewares := xarray.RemoveDuplicates(
		strings.Split(cfg.Middleware.Web, ","),
	)

	rpcRouter := r.Group(func(r chi.Router) {
		r.Use(middleware.GetMiddlewares(rpcMiddlewares...)...)
	})
	webRouter := r.Group(func(r chi.Router) {
		r.Use(middleware.GetMiddlewares(webMiddlewares...)...)
	})
	return &ServeMux{
		Router:    r,
		rpcRouter: rpcRouter,
		webRouter: webRouter,
		info: &serverInfo{
			address:    address,
			attributes: map[string]string{},
		},

		acceptHeaders: strings.Split(cfg.AcceptHeader, ","),
		outHeaders:    strings.Split(cfg.OutHeader, ","),
		outTrailers:   strings.Split(cfg.OutTrailer, ","),
	}
}

func (s *ServeMux) RpcHandle(meth, path string, f HandlerFunc) {
	s.rpcRouter.MethodFunc(meth, path, func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = metadata.WithStreamContext(ctx)
		ctx = metadata.WithInContext(ctx, s.extractInMetadata(r))
		ctx = peer.PeerWithContext(ctx, s.getPeer(r))
		r = r.WithContext(ctx)
		res, err := f(w, r)
		if err != nil {
			s.errorHandler(w, r, err)
			return
		}
		s.successHandler(w, r, res.(proto.Message))
	})
}

func (s *ServeMux) RawHandle(meth, path string, h http.HandlerFunc) {
	s.webRouter.MethodFunc(meth, path, h)
}

func (s *ServeMux) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stopped {
		return errors.New("server had already stopped")
	}
	if s.started {
		return errors.New("server had already serve")
	}
	s.started = true
	lis, err := net.Listen("tcp", s.info.address)
	if err != nil {
		return err
	}
	s.info.address = lis.Addr().String()
	s.listener = lis
	s.svr = &http.Server{
		Handler: s,
	}
	return nil
}

func (s *ServeMux) Serve() error {
	return s.svr.Serve(s.listener)
}

func (s *ServeMux) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.svr.Shutdown(ctx)
}

func (s *ServeMux) Info() ServerInfo {
	return s.info
}

func (s *ServeMux) extractInMetadata(r *http.Request) metadata.MD {
	var md = metadata.New(nil)
	for _, item := range s.acceptHeaders {
		vals := r.Header.Values(item)
		if vals == nil {
			continue
		}
		md.Append(item, vals...)
	}

	for key, vals := range r.Header {
		if strings.HasPrefix(key, MetadataHeaderPrefix) {
			md.Append(key[len(MetadataHeaderPrefix):], vals...)
		}
	}
	return md
}

func (s *ServeMux) getPeer(r *http.Request) *peer.Peer {
	ip, portStr, _ := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	port, _ := strconv.Atoi(portStr)
	clientIP := r.Header.Get("X-Forwarded-For")
	clientIP = strings.TrimSpace(strings.Split(clientIP, ",")[0])
	if clientIP == "" {
		clientIP = strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	}
	if clientIP != "" {
		ip = clientIP
	}
	return &peer.Peer{
		Addr: &net.TCPAddr{
			IP:   net.ParseIP(ip),
			Port: port,
		},
		LocalAddr: s.listener.Addr(),
		Protocol:  "http",
		IsRest:    true,
	}
}

func (s *ServeMux) errorHandler(w http.ResponseWriter, r *http.Request, err error) {
	ctx := r.Context()
	outbound := marshaler.OutboundFromContext(ctx)

	// return Internal when Marshal failed
	const fallback = `{"code": 13, "message": "failed to marshal error message"}`

	st := status.FromError(err)
	pb := st.Status()

	w.Header().Del("Trailer")
	w.Header().Del("Transfer-Encoding")

	contentType := outbound.ContentType(pb)
	w.Header().Set("Content-Type", contentType)

	if st.IsCode(code.Code_UNAUTHENTICATED) {
		w.Header().Set("WWW-Authenticate", st.Message())
	}

	buf, merr := outbound.Marshal(pb)
	if merr != nil {
		logger.Errorf("failed to marshal error message %q: %v", st, merr)
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := io.WriteString(w, fallback); err != nil {
			logger.Errorf("failed to write response: %v", err)
		}
		return
	}

	header, _ := metadata.FromHeaderCtx(ctx)
	trailerHeader, _ := metadata.FromTrailerCtx(ctx)

	s.handleResponseHeader(w, header)

	doForwardTrailers := s.requestAcceptsTrailers(r)

	if doForwardTrailers && trailerHeader.Len() > 0 {
		s.handleForwardResponseTrailerHeader(w, trailerHeader)
		w.Header().Set("Transfer-Encoding", "chunked")
	}

	w.WriteHeader(int(st.HttpCode()))
	if _, err := w.Write(buf); err != nil {
		logger.Errorf("Failed to write response: %v", err)
	}

	if doForwardTrailers && trailerHeader.Len() > 0 {
		s.handleForwardResponseTrailer(w, trailerHeader)
	}
}

func (s *ServeMux) successHandler(w http.ResponseWriter, r *http.Request, resp proto.Message) {
	ctx := r.Context()

	outbound := marshaler.OutboundFromContext(ctx)
	contentType := outbound.ContentType(resp)
	w.Header().Set("Content-Type", contentType)

	buf, err := outbound.Marshal(resp)
	if err != nil {
		logger.Infof("Marshal error: %v", err)
		s.errorHandler(w, r, err)
		return
	}

	header, _ := metadata.FromHeaderCtx(ctx)
	trailerHeader, _ := metadata.FromTrailerCtx(ctx)

	s.handleResponseHeader(w, header)

	doForwardTrailers := s.requestAcceptsTrailers(r)

	if doForwardTrailers && trailerHeader.Len() > 0 {
		s.handleForwardResponseTrailerHeader(w, trailerHeader)
		w.Header().Set("Transfer-Encoding", "chunked")
	}

	if _, err = w.Write(buf); err != nil {
		logger.Infof("Failed to write response: %v", err)
	}

	if doForwardTrailers && trailerHeader.Len() > 0 {
		s.handleForwardResponseTrailer(w, trailerHeader)
	}
}

func (s *ServeMux) handleResponseHeader(w http.ResponseWriter, md metadata.MD) {
	for k, vs := range md {
		if h, ok := s.outgoingHeaderMatcher(k); ok {
			for _, v := range vs {
				w.Header().Add(h, v)
			}
		}
	}
}

func (s *ServeMux) outgoingHeaderMatcher(key string) (string, bool) {
	for _, item := range s.outHeaders {
		if item == key {
			return key, true
		}
	}
	return fmt.Sprintf("%s%s", MetadataHeaderPrefix, key), true
}

func (s *ServeMux) requestAcceptsTrailers(req *http.Request) bool {
	te := req.Header.Get("TE")
	return strings.Contains(strings.ToLower(te), "trailers")
}

func (s *ServeMux) handleForwardResponseTrailerHeader(w http.ResponseWriter, md metadata.MD) {
	for k := range md {
		if h, ok := s.outgoingTrailerMatcher(k); ok {
			w.Header().Add("Trailer", h)
		}
	}
}

func (s *ServeMux) handleForwardResponseTrailer(w http.ResponseWriter, md metadata.MD) {
	for k, vs := range md {
		if h, ok := s.outgoingTrailerMatcher(k); ok {
			for _, v := range vs {
				w.Header().Add(h, v)
			}
		}
	}
}

func (s *ServeMux) outgoingTrailerMatcher(key string) (string, bool) {
	for _, item := range s.outTrailers {
		if item == key {
			return key, true
		}
	}
	return fmt.Sprintf("%s%s", MetadataTrailerPrefix, key), true
}
