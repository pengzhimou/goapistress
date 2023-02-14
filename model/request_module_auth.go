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

func setTKExHeader(req *http.Request, apiAccount string, apiToken string, reqPath string) error {
	// fmt.Printf("===TKExAuthData apiAccount:%s apiToken:%s reqPath: %s\n", apiAccount, apiToken, reqPath)
	localTimeStamp := time.Now().Unix()
	// t, _ := time.ParseInLocation("2022-10-08 15:23:00", time.Now().Format("2006-01-02 15:04:05"), time.Local)
	// localTimeStamp := t.Unix()
	localTime := strconv.FormatInt(localTimeStamp, 10)
	sigStr := computeSha256(apiToken + localTime + reqPath)
	req.Header.Set("TKEX-Api-Account", apiAccount)
	req.Header.Set("TKEX-Api-TimeStamp", localTime)
	req.Header.Set("TKEX-Api-Sig", sigStr)
	// fmt.Printf("===TKExHeader TKEX-Api-Account: %s TKEX-Api-TimeStamp:%s TKEX-Api-Sig:%s \n", apiAccount, localTime, sigStr)
	return nil
}

func (r *RequestForm) SetAuth(req *http.Request) {
	switch r.AuthType {
	case authBasic:
		r.Headers["username"] = r.AuthData["username"]
		r.Headers["password"] = r.AuthData["password"]
	case authTkexToken:
		setTKExHeader(req, r.AuthData["apiAccount"], r.AuthData["apiToken"], req.URL.Path)
	}
}
