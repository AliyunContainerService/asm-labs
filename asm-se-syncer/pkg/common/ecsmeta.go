package common

import (
	"encoding/json"
	"errors"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const (
	aliyunECSMetaBaseURL = "http://100.100.100.200/latest/meta-data/"
	aliyunECSMetaRamURL  = aliyunECSMetaBaseURL + "ram/security-credentials/"
	expirationTimeFormat = "2006-01-01T01:01:01Z"
)

type SecurityTokenResult struct {
	AccessKeyId     string
	AccessKeySecret string
	Expiration      string
	SecurityToken   string
	Code            string
	LastUpdated     string
}

func getEcsMetaRamRoleSecurityCredential() (result []byte, err error) {
	client := http.Client{
		Timeout: time.Second * 3,
	}
	var respList *http.Response
	respList, err = client.Get(aliyunECSMetaRamURL)
	if err != nil {
		log.Errorf("failed to get ram roles: %v", err)
		return nil, err
	}
	defer respList.Body.Close()
	var body []byte
	body, err = ioutil.ReadAll(respList.Body)
	if err != nil {
		log.Errorf("failed to parse ram roles: %v", err)
		return nil, err
	}
	if body == nil {
		log.Errorf("failed to parse ram roles: %v", err)
		return nil, errors.New("failed to parse ram roles with no body data")
	}

	bodyStr := string(body)
	bodyStr = strings.TrimSpace(bodyStr)
	roles := strings.Split(bodyStr, "\n")
	if roles == nil || len(roles) == 0 {
		log.Errorf("failed to parse ram roles: %v", err)
		return nil, errors.New("failed to parse ram roles with no role data")
	}
	role := roles[0]
	for _, ro := range roles {
		if strings.HasPrefix(ro, "Kubernetes") {
			role = ro
			break
		}
	}

	var respGet *http.Response
	respGet, err = client.Get(aliyunECSMetaRamURL + role)
	if err != nil {
		log.Errorf("failed to get token for role %s: %v", role, err)
		return nil, err
	}
	defer respGet.Body.Close()
	body, err = ioutil.ReadAll(respGet.Body)
	if err != nil {
		log.Errorf("failed to get token body for role %s: %v", role, err)
		return nil, err
	}
	return body, nil
}

func getSecurityTokenResult() (*SecurityTokenResult, error) {
	for tryTime := 0; tryTime < 3; tryTime++ {
		tokenResultBuffer, err := getEcsMetaRamRoleSecurityCredential()
		if err != nil {
			continue
		}
		var tokenResult SecurityTokenResult
		err = json.Unmarshal(tokenResultBuffer, &tokenResult)
		if err != nil {
			log.Errorf("failed to unmarshal token %s: %v", string(tokenResultBuffer), err)
			continue
		}
		if strings.ToLower(tokenResult.Code) != "success" {
			tokenResult.AccessKeySecret = "**********"
			tokenResult.SecurityToken = "**********"
			log.Errorf("failed to get successful code: %v", tokenResult)
			continue
		}

		return &tokenResult, nil
	}
	return nil, errors.New("failed to get security token")
}
