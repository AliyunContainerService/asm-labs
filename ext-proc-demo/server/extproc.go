package extproc

import (
	"encoding/json"
	"io"

	configPb "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	extProcPb "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	klog "k8s.io/klog/v2"
)

func NewServer() *Server {
	return &Server{}
}

// Server implements the Envoy external processing server.
// https://www.envoyproxy.io/docs/envoy/latest/api-v3/service/ext_proc/v3/external_processor.proto
type Server struct{}

func (s *Server) Process(srv extProcPb.ExternalProcessor_ProcessServer) error {
	klog.Infof("Processing")
	ctx := srv.Context()
	for {
		select {
		case <-ctx.Done():
			klog.Infof("context done")
			return ctx.Err()
		default:
		}

		req, err := srv.Recv()
		if err == io.EOF {
			// envoy has closed the stream. Don't return anything and close this stream entirely
			return nil
		}
		if err != nil {
			return status.Errorf(codes.Unknown, "cannot receive stream request: %v", err)
		}

		// build response based on request type
		resp := &extProcPb.ProcessingResponse{}
		switch v := req.Request.(type) {
		case *extProcPb.ProcessingRequest_RequestHeaders:
			klog.Infof("Got RequestHeaders")
			h := req.Request.(*extProcPb.ProcessingRequest_RequestHeaders)
			resp = handleRequestHeaders(h)
		case *extProcPb.ProcessingRequest_RequestBody:
			klog.Infof("Got RequestBody (not currently handled)")

		case *extProcPb.ProcessingRequest_RequestTrailers:
			klog.Infof("Got RequestTrailers (not currently handled)")

		case *extProcPb.ProcessingRequest_ResponseHeaders:
			klog.Infof("Got ResponseHeaders")
			h := req.Request.(*extProcPb.ProcessingRequest_ResponseHeaders)
			resp = handleResponseHeaders(h)

		case *extProcPb.ProcessingRequest_ResponseBody:
			klog.Infof("Got ResponseBody (not currently handled)")

		case *extProcPb.ProcessingRequest_ResponseTrailers:
			klog.Infof("Got ResponseTrailers (not currently handled)")

		default:
			klog.Infof("Unknown Request type %v", v)
		}

		klog.Infof("Sending ProcessingResponse: %+v", resp)
		if err := srv.Send(resp); err != nil {
			klog.Infof("send error %v", err)
			return err
		}

	}

}

func headerModifiers(key string, in *extProcPb.HttpHeaders) []*configPb.HeaderValueOption {
	modifiers := []*configPb.HeaderValueOption{}
	value := ""
	for _, header := range in.Headers.Headers {
		if header.Key == key {
			klog.Infof("found header-modifier, setting header, modifier: %s", string(header.RawValue))
			value = string(header.RawValue)
			break
		}
	}
	if value != "" {
		unmarshaled := map[string]string{}
		if err := json.Unmarshal([]byte(value), &unmarshaled); err != nil {
			klog.Errorf("error unmarshalling header-modifier: %s", err)
			return modifiers
		}
		for k, v := range unmarshaled {
			modifiers = append(modifiers, &configPb.HeaderValueOption{
				Header: &configPb.HeaderValue{
					Key:      k,
					RawValue: []byte(v),
				},
			})
		}
	}
	return modifiers
}

func handleRequestHeaders(req *extProcPb.ProcessingRequest_RequestHeaders) *extProcPb.ProcessingResponse {
	klog.Infof("handle request headers: %+v\n", req)

	resp := &extProcPb.ProcessingResponse{
		Response: &extProcPb.ProcessingResponse_RequestHeaders{
			RequestHeaders: &extProcPb.HeadersResponse{
				Response: &extProcPb.CommonResponse{
					HeaderMutation: &extProcPb.HeaderMutation{
						SetHeaders: []*configPb.HeaderValueOption{
							{
								Header: &configPb.HeaderValue{
									Key:      "x-ext-proc-header",
									RawValue: []byte("hello-to-asm"),
								},
							},
						},
					},
				},
			},
		},
	}
	resp.Response.(*extProcPb.ProcessingResponse_RequestHeaders).RequestHeaders.Response.HeaderMutation.SetHeaders = append(resp.Response.(*extProcPb.ProcessingResponse_RequestHeaders).RequestHeaders.Response.HeaderMutation.SetHeaders, headerModifiers("request-header-modifier", req.RequestHeaders)...)
	return resp
}

func handleResponseHeaders(req *extProcPb.ProcessingRequest_ResponseHeaders) *extProcPb.ProcessingResponse {
	klog.Infof("handle response headers: %+v\n", req)

	resp := &extProcPb.ProcessingResponse{
		Response: &extProcPb.ProcessingResponse_ResponseHeaders{
			ResponseHeaders: &extProcPb.HeadersResponse{
				Response: &extProcPb.CommonResponse{
					HeaderMutation: &extProcPb.HeaderMutation{
						SetHeaders: []*configPb.HeaderValueOption{
							{
								Header: &configPb.HeaderValue{
									Key:      "x-ext-proc-header",
									RawValue: []byte("hello-from-asm"),
								},
							},
						},
					},
				},
			},
		},
	}

	return resp
}
