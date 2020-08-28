package rostering

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Khan/clever-repartee/pkg/tripperware"

	"go.uber.org/zap"

	"github.com/pkg/errors"
)

func GetCleverToken(
	logger *zap.Logger,
	districtID string,
	isMAP bool,
) (string, error) {
	req, err := http.NewRequest(
		"GET",
		"https://clever.com/oauth/tokens?owner_type=district&district="+districtID,
		nil,
	)
	if err != nil {
		return "", err
	}

	var creds string

	if isMAP {
		creds = os.ExpandEnv("${MAP_CLEVER_ID}:${MAP_CLEVER_SECRET}")
	} else {
		creds = os.ExpandEnv("${CLEVER_ID}:${CLEVER_SECRET}")
	}

	if strings.HasPrefix(creds, ":") || strings.HasSuffix(creds, ":") {
		return "", errors.New(
			"all environment variables must be set including " +
				"${MAP_CLEVER_ID} ${MAP_CLEVER_SECRET} ${CLEVER_ID} and ${CLEVER_SECRET}",
		)
	}
	req.Header.Set(
		"Authorization",
		"Basic "+base64.StdEncoding.EncodeToString([]byte(creds)),
	)

	pesterClient := tripperware.NewLoggedRetryHTTPClient(logger)

	resp, err := pesterClient.Do(req)
	if err != nil {
		return "", err
	}

	if !IsHTTPSuccess(resp.StatusCode) {
		resp.Body.Close()
		return "", fmt.Errorf(
			"HTTP %d Error for Clever Request /oauth/tokens?owner_type=district&district=%s",
			resp.StatusCode,
			districtID,
		)
	}

	if resp == nil {
		return "", nil
	}

	tokenResp := &TokenResponse{}
	if resp.Body == nil {
		return "", nil
	}

	dec := json.NewDecoder(resp.Body)
	dec.UseNumber()
	err = dec.Decode(tokenResp)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	if tokenResp.Data != nil && len(tokenResp.Data) != 0 {
		return tokenResp.Data[0].AccessToken, nil
	}
	return "", nil
}

type TokenResponse struct {
	Data []Data `json:"data"`
}
type Owner struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}
type Data struct {
	ID          string    `json:"id"`
	Created     time.Time `json:"created"`
	Owner       Owner     `json:"owner"`
	AccessToken string    `json:"access_token"`
	Scopes      []string  `json:"scopes"`
}
