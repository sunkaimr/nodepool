package webhook

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"net/http"
	"nodepool/controllers"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	CertFile = "-----BEGIN CERTIFICATE-----\nMIICszCCAZsCFDBSq45c+6BOU10TSf6C6W67mSGKMA0GCSqGSIb3DQEBCwUAMBYx\nFDASBgNVBAMMC25vZGVwb29sLmlvMB4XDTIyMDQwNjA5MTExM1oXDTIzMDQwNjA5\nMTExM1owFjEUMBIGA1UEAwwLbm9kZXBvb2wuaW8wggEiMA0GCSqGSIb3DQEBAQUA\nA4IBDwAwggEKAoIBAQChaF+gLMNIqJ327oh7cjgLN+5rib1LWaNIRwf2sYNHkgZe\nL3BamYqyA8iSE6cVXOLGEdGkLVwtMPGw+7IKuxcnGbSEiHrro5Ieca0yjDL0ZoxA\nHdWN4EJTQU0xoPeiLgPPZpym5uIK3B9vc09Zh7Vtp4E0HpRUxkJzCVuG7PoQlQs8\nvcVcPiC76jrlc2vsx5G4pRa/6Uuc8RTzz3fCRA9RglWjbOK9/mLgIUmuS4WIcM+j\ntRpKkJrsauM/2H6BRGkC3VZPA1+HtWO3kXuleIUdosWv9+nxfrYtNRKXFhoO1jlZ\nE0Igv00hjYO/cSSWmSyQAc53Y31oZEXf6lQNzI6PAgMBAAEwDQYJKoZIhvcNAQEL\nBQADggEBAMZm2f68VACUalJztgCi+rnqKhYaPoVRjlM3J0OB8NFAy2UWDgdO02uy\nxbkrJPctrE3D+zWTivUjSO1FbXp71TrhBekJDlKfWcWB650dVUo0WNpAiARdH28H\nSBPi/GjGCpbv/AikbjtjeCzs4WVHSnyOU8aU5aijzGbRzg2A/G72P8Xnj/a75Zyd\n5XexB/F1FLmQI0ijBm/WFkLzFVMGctPVFQVJRnauy8aI5YUdOy6arZ3Tp16G+qXY\n7eiydVlQ1JsBuNaFu7EMNuP7T+jE2tnfsXrrLQs7HpTsJKeuzsX/z5XdFhC1PgCP\nK870iyVcEiW0PQZpFD/ihNzIhOPw+rA=\n-----END CERTIFICATE-----"

	KeyFile  = "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEAoWhfoCzDSKid9u6Ie3I4Czfua4m9S1mjSEcH9rGDR5IGXi9w\nWpmKsgPIkhOnFVzixhHRpC1cLTDxsPuyCrsXJxm0hIh666OSHnGtMowy9GaMQB3V\njeBCU0FNMaD3oi4Dz2acpubiCtwfb3NPWYe1baeBNB6UVMZCcwlbhuz6EJULPL3F\nXD4gu+o65XNr7MeRuKUWv+lLnPEU8893wkQPUYJVo2zivf5i4CFJrkuFiHDPo7Ua\nSpCa7GrjP9h+gURpAt1WTwNfh7Vjt5F7pXiFHaLFr/fp8X62LTUSlxYaDtY5WRNC\nIL9NIY2Dv3EklpkskAHOd2N9aGRF3+pUDcyOjwIDAQABAoIBAG/5MYW0KBHK8DMz\nTbmeBmU5+wvddVXFrLHinRK2GSXYltRWQrKHnCFc4JL+UUZPtv7Ds9UapryvHKy0\nH9Kz8h7tBT+AUw4P3rmCES5k9qB4V8nPKKyRLFIHll7clY6ML5Z9UCW1PJFeHey6\naQzqSaH1t3uJz0t0cvrNPhK/aceAFEVk7KszLxCEoYdVhnVcVkjay/VWPNCrJPfo\ncm/Jw8oaxub+Bra/DBqdp4eYiaStA+FPO0nNdjIKp7cWZ6wV/7iBkj+OsaS7P0hv\nDpDYV0X1GCFlu1lMNYK69fOYsxun7fCojl31vtCNHsB8ij8xTYkX2Gjj8WEqVoh9\n0sZv2ZkCgYEAzox5geaMx1zFvIyNf2CgB1xczUtAEKkQvQBJ6r/yRdqPcsoh459x\nikbzaQw0e7rqc+kwpzOPIlCyiXJU4OYfrmJbwcsg2Bl4e3ZfhDhxkDOezV77eEvm\neia3jaPCDTnaxYui4+oKB9U7RmB71kUi7sS/KJevNOuGlTtEvZszDN0CgYEAyA0s\nFEvXL5rRh80pHzxYkxwJGxjsrIwI/O+t4ECIMXbpa75pkTtV77EYG7y9X6RIJx0K\n9rf6TaHEsdWtj2QwjIf7mVFD4f4TLCEE+LdDTlEYQcWSFNMTjEknu1FEcBaI4YwY\n8g/VRABblzMk4Zog8A9T6huz+kJ/6v6/UvqlLFsCgYEAghCZb0B0BBKafenwLHb3\nLsttsOUi+ZrM7IdjBI1MjcpcrIc2ofTEdbPKEata3VNN0iHUvmVMS+qPEthJNLoU\n1yYe68Dy9MHNScm3yjYU5R5scJzQM+dvwhnhWjL1vohhCCavM2AsYtRWmDGnqb0t\nizefvHsQHH336L9CwTcbBY0CgYAqNC0ycvWIw36kybF9N3vwPR/mqZF0rW5P/jiO\ncM7KsK1534fh7cSpdpEBeQXyoXPfXI8tkY6qxg/6/HtLHvXnD+ESbUSG7tUYoDau\nSetXIGCfr5Cr+APNurk5GWH4y6hA/Q9eMdzqJbEs6stDFQMR4gnv/7wudc0KxIeU\nd/BUtQKBgQCe9K/P2x9XCsvysyPm3hEzZI6C/59AzNCcs9r4aZubJwv4lDFaz/a3\nS/Sj7pd5TMHEeazCr6XXuOcmzIY/8sRkvZFH7r+DZ2NW4mGvGPUpjm1tXL1qWw+4\no/uC0CNZoP3ZYS2YQf+xIrbnL7YefZnUluAMXkzLs+ptQV85a2wj+g==\n-----END RSA PRIVATE KEY-----"

	)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()
)

type Server struct {
	server *http.Server
}

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

func NewServer(addr string, port int) *Server {
	//tlsCertKey, err := tls.X509KeyPair([]byte(CertFile), []byte(KeyFile))
	tlsCertKey, err := tls.LoadX509KeyPair("/root/test/server.crt", "/root/test/server.key")
	if err != nil {
		panic(err)
	}

	s := &Server{
		server: &http.Server{
			Addr:      fmt.Sprintf("%s:%d", addr, port),
			TLSConfig: &tls.Config{Certificates: []tls.Certificate{tlsCertKey}},
		},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/mutating", s.mutatingHandle)
	s.server.Handler = mux
	return s
}

func (s *Server) Start() {
	go func() {
		err := s.server.ListenAndServeTLS("", "")
		if err != nil {
			panic(err)
		}
	}()
}

func (s *Server) mutatingHandle(w http.ResponseWriter, r *http.Request) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}
	if len(body) == 0 {
		log.Log.Error(fmt.Errorf("empty body"), "")
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		log.Log.Error(fmt.Errorf("Content-Type=%s, expect application/json", contentType), "")
		http.Error(w, "invalid Content-Type, expect `application/json`", http.StatusUnsupportedMediaType)
		return
	}

	var admissionResponse *v1beta1.AdmissionResponse
	ar := v1beta1.AdmissionReview{}
	if _, _, err := deserializer.Decode(body, nil, &ar); err != nil {
		log.Log.Error(err, "Can't decode body")
		admissionResponse = &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	admissionResponse = s.mutating(&ar)

	admissionReview := v1beta1.AdmissionReview{}
	if admissionResponse != nil {
		admissionReview.Response = admissionResponse
		admissionReview.APIVersion = ar.APIVersion
		admissionReview.Kind = ar.Kind
		if ar.Request != nil {
			admissionReview.Response.UID = ar.Request.UID
		}
	}

	resp, err := json.Marshal(admissionReview)
	if err != nil {
		log.Log.Error(err, "Can't encode response")
		http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
	}

	if _, err := w.Write(resp); err != nil {
		log.Log.Error(err, "Can't write response: %v")
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
	}

}

// main mutation process
func (s *Server) mutating(ar *v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	req := ar.Request
	pod := corev1.Pod{}
	patchBytes := []byte{}
	err := errors.New("")
	resp := &v1beta1.AdmissionResponse{
		Allowed: false,
		Result: &metav1.Status{},
	}

	log.Log.Info(fmt.Sprintf("AdmissionReview for Kind=%v, Namespace=%v Name=%v UID=%v patchOperation=%v UserInfo=%v",
		req.Kind, req.Namespace, req.Name, req.UID, req.Operation, req.UserInfo))

	switch req.Kind.Kind {
	case "Pod":
		if err = json.Unmarshal(req.Object.Raw, &pod); err != nil {
			log.Log.Error(err, "Could not unmarshal raw object: %v")
			resp.Result.Message = err.Error()
			return resp
		}
	default:
		log.Log.Info(fmt.Sprintf("no need Admission resource of kind %s", req.Kind.Kind))
	}

	patchBytes, err = patchPod(ar)
	if err != nil {
		resp.Result.Message = err.Error()
		return resp
	}
	log.Log.Info(fmt.Sprintf("AdmissionResponse: patch=%v", string(patchBytes)))

	resp.Allowed = true
	resp.Patch = patchBytes
	resp.PatchType = func() *v1beta1.PatchType {
		pt := v1beta1.PatchTypeJSONPatch
		return &pt
	}()
	return resp
}

func patchPod(ar *v1beta1.AdmissionReview) ([]byte, error) {
	var patch []patchOperation

	if !controllers.InclusionExceptionNs(ar.Request.Namespace) {
		values := make(map[string]string)
		values[controllers.LableNodePoolKey] = ar.Request.Namespace
		patch = append(patch, patchOperation{
			Op:    "add",
			Path:  "/spec/nodeSelector",
			Value: values,
		})
	}

	return json.Marshal(patch)
}
