package model

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

const (
	authBasic      = "basic"
	authTkexToken  = "tkextoken"
	authForceHttps = "forcehttps"
)

func computeSha256(str string) string {

	h := sha256.New()
	_, _ = h.Write([]byte(str))
	sha256Byte := h.Sum(nil)

	return fmt.Sprintf("%x", sha256Byte)
}

func setAuth(req *http.Request, apiAccount string, apiToken string, reqPath string) error {
	fmt.Println("===setAuth reqPath: %s", reqPath)
	localTimeStamp := time.Now().Unix()
	// t, _ := time.ParseInLocation("2022-10-08 15:23:00", time.Now().Format("2006-01-02 15:04:05"), time.Local)
	// localTimeStamp := t.Unix()

	localTime := strconv.FormatInt(localTimeStamp, 10)

	sigStr := computeSha256(apiToken + localTime + reqPath)

	req.Header.Set("TKEX-Api-Account", apiAccount)
	req.Header.Set("TKEX-Api-TimeStamp", localTime)
	req.Header.Set("TKEX-Api-Sig", sigStr)

	return nil
}

// func (c *Connection) TokenInjection() *Connection {
// 	if c.AuthType == "httptoken" {
// 		uriSlc := strings.Split(c.UrlF.RequestURI(), "?")
// 		pureURI := uriSlc[0]
// 		setAuth(c.Req, TestConfig.MBAuth.APIUser, TestConfig.MBAuth.APIToken, pureURI)
// 		// setAuth(c.Req, TestConfig.MBAuth.APIUser, TestConfig.MBAuth.APIToken, c.UrlF.Path)
// 		klog.Info(c.Uuid, " Request Header After Injection is: ", c.Req.Header)
// 	} else {
// 		klog.Info(c.Uuid, " Will Not Inject Token to Header ", c.Req.Header)
// 	}
// 	return c
// }

func (r *RequestForm) setAuth() {
	switch r.AuthType {
	case authBasic:
		r.Headers["username"] = r.AuthData["username"]
		r.Headers["password"] = r.AuthData["password"]
	case authTkexToken:

	case authForceHttps:
	}
}
