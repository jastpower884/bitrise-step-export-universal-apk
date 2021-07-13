package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/bitrise-io/go-steputils/stepconf"
	"github.com/bitrise-io/go-steputils/tools"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-steplib/bitrise-step-export-universal-apk/apkexporter"
	"github.com/bitrise-steplib/bitrise-step-export-universal-apk/bundletool"
	"github.com/bitrise-steplib/bitrise-step-export-universal-apk/filedownloader"
)

// Config is defining the input arguments required by the Step.
type Config struct {
	DeployDir         string `env:"BITRISE_DEPLOY_DIR"`
	AABPathList       string `env:"aab_path_list,required"`
	KeystoreURL       string `env:"keystore_url"`
	KeystotePassword  string `env:"keystore_password"`
	KeyAlias          string `env:"keystore_alias"`
	KeyPassword       string `env:"private_key_password"`
	BundletoolVersion string `env:"bundletool_version"`
}

func main() {
	var config Config
	if err := stepconf.Parse(&config); err != nil {
		failf("Error: %s \n", err)
	}
	stepconf.Print(config)
	fmt.Println()

	bundletoolTool, err := bundletool.New(config.BundletoolVersion, filedownloader.New(http.DefaultClient))
	if err != nil {
		failf("Failed to initialize bundletool: %s \n", err)
	}
	log.Infof("bundletool path created at: %s", bundletoolTool.Path())

	exporter := apkexporter.New(bundletoolTool, filedownloader.New(http.DefaultClient))

	aabPathList := parseAppList(config.AABPathList)

	apkPaths := make([]string, 0)
	for _, aabPath := range aabPathList {
		keystoreCfg := parseKeystoreConfig(config)
		apkPath, err := exporter.ExportUniversalAPK(aabPath, config.DeployDir, keystoreCfg)
		if err != nil {
			failf("Failed to export apk, error: %s \n", err)
		}

		//if err = tools.ExportEnvironmentWithEnvman("BITRISE_APK_PATH", apkPath); err != nil {
		aabPathList = append(apkPaths, aabPath)
		log.Donef("Success! APK exported to: %s", apkPath)
	}

	joinedAPKOutputPaths := strings.Join(aabPathList, "|")

	if err = tools.ExportEnvironmentWithEnvman("BITRISE_APK_PATH_LIST", joinedAPKOutputPaths); err != nil {
		failf("Failed to export BITRISE_APK_PATH, error: %s \n", err)
	}
	os.Exit(0)
}

func parseKeystoreConfig(config Config) *bundletool.KeystoreConfig {
	if config.KeystoreURL == "" ||
		config.KeystotePassword == "" ||
		config.KeyAlias == "" ||
		config.KeyPassword == "" {
		return nil
	}

	return &bundletool.KeystoreConfig{
		Path:               strings.TrimSpace(config.KeystoreURL),
		KeystorePassword:   config.KeystotePassword,
		SigningKeyAlias:    config.KeyAlias,
		SigningKeyPassword: config.KeyPassword}
}

func failf(s string, a ...interface{}) {
	log.Errorf(s, a...)
	os.Exit(1)
}

func splitElements(list []string, sep string) (s []string) {
	for _, e := range list {
		s = append(s, strings.Split(e, sep)...)
	}
	return
}

func parseAppList(list string) (apps []string) {
	list = strings.TrimSpace(list)
	if len(list) == 0 {
		return nil
	}

	s := []string{list}
	for _, sep := range []string{"\n", `\n`, "|"} {
		s = splitElements(s, sep)
	}

	for _, app := range s {
		app = strings.TrimSpace(app)
		if len(app) > 0 {
			apps = append(apps, app)
		}
	}
	return
}
