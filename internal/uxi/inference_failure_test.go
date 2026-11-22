package uxi

import (
 "context"
 "errors"
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"
)

func TestInferenceFailureKinds(t *testing.T) {
 cases := []struct{ status int; want string }{
  {http.StatusRequestEntityTooLarge,"request_rejected_possible_context_limit"},
  {http.StatusUnprocessableEntity,"request_rejected_possible_context_limit"},
  {http.StatusTooManyRequests,"capacity_limited"},
  {http.StatusGatewayTimeout,"timeout"},
  {http.StatusServiceUnavailable,"provider_unavailable"},
  {http.StatusUnauthorized,"provider_http_error"},
 }
 for _, tc := range cases {
  got := classifyInferenceHTTP(tc.status)
  if inferenceFailureKind(got)!=tc.want { t.Errorf("status %d: got %s want %s",tc.status,inferenceFailureKind(got),tc.want) }
 }
 if got:=inferenceFailureKind(classifyInferenceTransport(context.DeadlineExceeded));got!="timeout"{t.Fatal(got)}
 if got:=inferenceFailureKind(classifyInferenceTransport(context.Canceled));got!="cancelled"{t.Fatal(got)}
 if !errors.Is(classifyInferenceTransport(context.Canceled),context.Canceled){t.Fatal("cause lost")}
}
func TestLocalInferenceHTTPErrorIsNotOffline(t *testing.T) {
 server:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  if r.URL.Path!="/v1/chat/completions"{t.Errorf("unexpected path %s",r.URL.Path)}
  w.WriteHeader(http.StatusUnprocessableEntity)
  _,_ = w.Write([]byte("sensitive provider payload must not be disclosed"))
 }))
 defer server.Close()
 client:=NewLocalInferenceClient(server.URL+"/v1","aetherion:latest")
 _,err:=client.Complete(context.Background(),"system",[]ChatMessage{{Role:"user",Content:"hello"}})
 if inferenceFailureKind(err)!="request_rejected_possible_context_limit"{t.Fatalf("wrong failure: %v",err)}
 if strings.Contains(err.Error(),"sensitive provider"){t.Fatal("provider payload leaked")}
 answer:=degradedAnswer(nil,false,err)
 if strings.Contains(answer,"endpoint is not available")||!strings.Contains(answer,"does not prove the model is offline"){t.Fatalf("misleading degraded message: %s",answer)}
}
