package jablkodev

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

type EnvironmentVars struct {
	CorePort   int    // HTTP Port that Jablko is listening on: JABLKO_CORE_PORT
	JMODPort   int    // Port that the JMOD should listen on: JABLKO_MOD_PORT
	JMODKey    string // String used to authenticate the JMOD when calling Jablko routes: JABLKO_MOD_KEY
	JMODConfig string // Config string for JMOD: JABLKO_MOD_CONFIG
}

// Global storage of environment variables initialized by LoadJablkoEnv
var JABLKO_ENV EnvironmentVars

// Reads in environment variables set by Jablko.
// Throws an error for any invalid value
func ReadEnvironmentVars() (EnvironmentVars, error) {
	corePort, err := strconv.Atoi(os.Getenv("JABLKO_CORE_PORT"))
	if err != nil {
		return EnvironmentVars{}, fmt.Errorf("Error reading JABLKO_CORE_PORT: %v", err)
	}

	jmodPort, err := strconv.Atoi(os.Getenv("JABLKO_MOD_PORT"))
	if err != nil {
		return EnvironmentVars{}, fmt.Errorf("Error reading JABLKO_MOD_PORT: %v", err)
	}

	jmodKey := os.Getenv("JABLKO_MOD_KEY")
	if len(jmodKey) == 0 {
		return EnvironmentVars{}, fmt.Errorf("JABLKO_MOD_KEY not defined: %v", err)
	}

	jmodConfig := os.Getenv("JABLKO_MOD_CONFIG")
	if len(jmodConfig) == 0 {
		return EnvironmentVars{}, fmt.Errorf("JABLKO_MOD_CONFIG not defined: %v", err)
	}

	return EnvironmentVars{corePort, jmodPort, jmodKey, jmodConfig}, nil
}

// Sets package global JABLKO_ENV.
// This function must be called first before using other utilities in this module.
// Returns an error if ReadEnvironmentVars fails.
// Calling JMODS should terminate if this call fails.
func LoadJablkoEnv() error {
	vars, err := ReadEnvironmentVars()
	if err != nil {
		return err
	}

	JABLKO_ENV = vars
	return nil
}

func GetJablkoCorePort() int {
	return JABLKO_ENV.CorePort
}

func GetJablkoModPort() int {
	return JABLKO_ENV.JMODPort
}

func GetJablkoModConfig() string {
	return JABLKO_ENV.JMODConfig
}

func NewJablkoRequest(method string, url string, body io.Reader) (*http.Request, error) {
	r, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	r.Header.Set("JABLKO_MOD_PORT", strconv.Itoa(JABLKO_ENV.JMODPort))
	r.Header.Set("JABLKO_MOD_KEY", JABLKO_ENV.JMODKey)

	return r, nil
}

func NewJablkoRequestWithContext(ctx context.Context, method string, url string, body io.Reader) (*http.Request, error) {
	r, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	r.Header.Set("JABLKO_MOD_PORT", strconv.Itoa(JABLKO_ENV.JMODPort))
	r.Header.Set("JABLKO_MOD_KEY", JABLKO_ENV.JMODKey)

	return r, nil
}

// This returns a byte slice that contains the content of the response body.
// This function is meant to be a simple wrapper for a common pattern.
func PostSimple(url string, contentType string, body io.Reader) ([]byte, error) {
	req, err := NewJablkoRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Bad status code: %v", resp.Status)
	}

	return ioutil.ReadAll(resp.Body)
}

// This returns a byte slice that contains the content of the response body.
// This function is meant to be a simple wrapper for a common pattern.
func GetSimple(url string) ([]byte, error) {
	req, err := NewJablkoRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Bad status code: %v", resp.Status)
	}

	return ioutil.ReadAll(resp.Body)
}
