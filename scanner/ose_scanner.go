/*
Copyright (C) 2016 Black Duck Software, Inc.
http://www.blackducksoftware.com/

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownership. The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied. See the License for the
specific language governing permissions and limitations
under the License.
*/

package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	APP_VERSION                   = "0.1"
	DEFAULT_BDS_SCANNER_BASE_DIR  = "/tmp/ocp-scanner"
	CLI_IMPL_JAR_FILE_NAME        = "scan.cli.impl-standalone.jar"
	BDS_SCANNER_BASE_DIR_VAR_NAME = "SCANNER_BASE_DIR"
	SCAN_CLI_JAR_NAME_VAR_NAME    = "SCAN_CLI_JAR_NAME"
	APP_HOME_VAR_NAME             = "APP_HOME"
)

type input struct {
	host        string
	port        string
	scheme      string
	username    string
	password    string
	imageId     string
	taggedImage string
	digest      string
}

var in input

func init() {
	flag.StringVar(&in.host, "h", "REQUIRED", "The hostname of the Black Duck Hub server.")
	flag.StringVar(&in.port, "p", "REQUIRED", "The port the Hub is communicating on")
	flag.StringVar(&in.scheme, "s", "REQUIRED", "The communication scheme [http,https].")
	flag.StringVar(&in.username, "u", "REQUIRED", "The Black Duck Hub user")
	flag.StringVar(&in.password, "w", "REQUIRED", "Password for the user. You're not prompted since this is automated run.")
	flag.StringVar(&in.imageId, "id", "REQUIRED", "Image ID")
	flag.StringVar(&in.taggedImage, "tag", "REQUIRED", "Tagged image name")
	flag.StringVar(&in.digest, "digest", "REQUIRED", "Digest")
}

// The flag package provides a default help printer via -h switch
var versionFlag *bool = flag.Bool("v", false, "Print the version number.")
var dumpFlag *bool = flag.Bool("d", false, "dumps extracted tar in /tmp.")

func scanImage(path string, imageId string, taggedImage string) {
	log.Println("Scanning " + taggedImage)

	img_arr := strings.Split(taggedImage, "@sha256:")
	img_name := img_arr[0]
	img_ps := img_arr[1]
	prefix := img_ps[:10]
	project := img_name

	appHomeDir := os.Getenv(APP_HOME_VAR_NAME)
	scanCliJarName := os.Getenv(SCAN_CLI_JAR_NAME_VAR_NAME)
	scanCliImplJarPath := filepath.Join(appHomeDir, "lib", "cache", CLI_IMPL_JAR_FILE_NAME)
	scanCliJarPath := filepath.Join(appHomeDir, "lib", scanCliJarName)

	log.Println("Processing project " + project + " with version " + prefix)
	log.Printf("Scan CLI Impl Jar: %s\n", scanCliImplJarPath)
	log.Printf("Scan CLI Jar: %s\n", scanCliJarPath)

	cmd := exec.Command("java",
		"-Xms512m",
		"-Xmx4096m",
		"-Done-jar.silent=true",
		"-Done-jar.jar.path="+scanCliImplJarPath,
		"-jar", scanCliJarPath,
		"--host", in.host,
		"--port", in.port,
		"--scheme", in.scheme,
		"--project", project,
		"--release", prefix,
		"--username", in.username,
		"--password", in.password,
		"-v",
		path)
	// if we print this out, we're going to print out the password - bad idea
	//log.Println(cmd.Args)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		log.Println(err.Error())
		return
	}

}

func writeContents(body io.ReadCloser, path string) (tarFilePath string, err error) {

	defer func() {
		body.Close()
	}()

	log.Println("Starting to write file contents to a tar file.")

	tarFilePath = fmt.Sprintf("%s.%s", path, "tar")
	log.Printf("Tar File Path: %s\n", tarFilePath)

	f, err := os.OpenFile(tarFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		fmt.Println("ERROR : opening file.")
		fmt.Println(err)
		return "", err
	}

	if _, err := io.Copy(f, body); err != nil {
		fmt.Println("ERROR : copying into file.")
		fmt.Println(err)
		return "", err
	}

	return tarFilePath, nil
}

func getHttpRequestResponse(client *httputil.ClientConn, httpMethod string, requestUrl string) (resp *http.Response, err error) {

	log.Printf("Making request: [%s] [%s]\n", httpMethod, requestUrl)
	req, err := http.NewRequest(httpMethod, requestUrl, nil)

	if err != nil {
		return nil, err
	}

	resp, err = client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("received status != 200 on resp OK: %s", resp.Status))
	}

	return resp, nil
}

func saveImageToTar(client *httputil.ClientConn, image string, path string) (tarFilePath string, err error) {

	exists, err := imageExists(client, image)
	if err != nil {
		return "", err
	}

	if !exists {
		return "", nil
	}

	os.MkdirAll(path, 0755)
	imageUrl := fmt.Sprintf("/images/%s/get", image)
	resp, err := getHttpRequestResponse(client, "GET", imageUrl)

	if err != nil {
		return "", err
	}

	return writeContents(resp.Body, path)

}

func imageExists(client *httputil.ClientConn, image string) (result bool, err error) {

	imageUrl := fmt.Sprintf("/images/%s/history", image)
	resp, err := getHttpRequestResponse(client, "GET", imageUrl)

	if err != nil {
		log.Printf("Error testing for image presence\n%s\n", err.Error())
		return false, err
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		log.Printf("Error reading image history\n%s\n", err.Error())
		return false, err
	} else if len(buf.Bytes()) == 0 {
		log.Printf("No data in image history\n")
	}

	return true, nil
}

func getScannerOutputDir() string {
	// NOTE: At this point we don't have a logger yet, so don't try and use it.


	// Check to see if we can get the env var for the base dir (we should always be able to)
	bdsScannerBaseDir := os.Getenv(BDS_SCANNER_BASE_DIR_VAR_NAME)

	// If for some reason we can't, make up a default, since this doesn't have any outside dependencies
	if bdsScannerBaseDir == "" {
		bdsScannerBaseDir = DEFAULT_BDS_SCANNER_BASE_DIR
	}

	// Make base out dir for any scanner output, scans, etc...
	scannerOutputDir := bdsScannerBaseDir
	os.MkdirAll(scannerOutputDir, os.ModeDir|os.ModePerm)

	fmt.Printf("Scanner Output Dir: %s\n", scannerOutputDir)
	return scannerOutputDir
}

func checkExpectedCmdlineParams() bool {
	// NOTE: At this point we don't have a logger yet, so don't try and use it.

	if *versionFlag {
		fmt.Println("Version:", APP_VERSION)
		return false
	}

	// These checks seem a little odd, I'd expect the flag parsing to be
	// able to handle most of this...
	if in.host == "REQUIRED" {
		fmt.Println("-h host is required")
		flag.PrintDefaults()
		return false
	}

	if in.port == "REQUIRED" {
		fmt.Println("-p port is required")
		flag.PrintDefaults()
		return false
	}

	if in.scheme == "REQUIRED" {
		fmt.Println("-s scheme is required")
		flag.PrintDefaults()
		return false
	}

	if in.username == "REQUIRED" {
		fmt.Println("-u username is required")
		flag.PrintDefaults()
		return false
	}

	if in.password == "REQUIRED" {
		fmt.Println("-w password is required")
		flag.PrintDefaults()
		return false
	}

	if in.imageId == "REQUIRED" {
		fmt.Println("-id image ID is required")
		flag.PrintDefaults()
		return false
	}

	if in.taggedImage == "REQUIRED" {
		fmt.Println("-tag Image tag is required")
		flag.PrintDefaults()
		return false
	}

	if in.digest == "REQUIRED" {
		fmt.Println("-digest Image digest is required")
		flag.PrintDefaults()
		return false
	}

	return true
}

func checkExpectedEnvVars() bool {
	// NOTE: At this point we don't have a logger yet, so don't try and use it.

	appHomeDir := os.Getenv(APP_HOME_VAR_NAME)
	scanCliJarName := os.Getenv(SCAN_CLI_JAR_NAME_VAR_NAME)
	errorMsgFmt := "%s env var is required to be set.\n"

	if appHomeDir == "" {
		fmt.Printf(errorMsgFmt, APP_HOME_VAR_NAME)
		return false
	}

	if scanCliJarName == "" {
		fmt.Printf(errorMsgFmt, SCAN_CLI_JAR_NAME_VAR_NAME)
		return false
	}

	return true
}

func healthy(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func validatePreCacheMode() {
	val := os.Getenv("BDS_LISTEN")
	if val != "9036" {
		// flag wasn't set, so not in pre-cache mode
		fmt.Println ("Precache flag missing. Configuring normal mode.")
		return
	}

	fmt.Println ("Operating in pre-cache mode.")

	http.HandleFunc("/health", healthy)      // set router
	err := http.ListenAndServe(":9036", nil) // set listen port

	if err != nil {
		log.Fatal("validatePreCacheMode: ", err)
	}
}

func main() {

	validatePreCacheMode()

	// check input arguments
	flag.Parse() // Scan the arguments list

	if !checkExpectedCmdlineParams() {
		return
	}

	if !checkExpectedEnvVars() {
		return
	}

	scannerOutputDir := getScannerOutputDir()

	// Route the log to a file
	logFile := fmt.Sprintf("%s/%s.log", scannerOutputDir, os.Args[0])
	fmt.Printf("Log File: %s\n", logFile)

	f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		msg := fmt.Sprintf("ERROR : can't create logfile.\n %v\n", err)
		log.Fatal(msg)
	}

	defer f.Close()
	log.SetOutput(io.MultiWriter(os.Stdout, f))

	log.Printf("APP_HOME env var: [%s]\n", os.Getenv(APP_HOME_VAR_NAME))

	c, err := net.Dial("unix", "/var/run/docker.sock")
	if err != nil {
		panic(err)
	}
	defer c.Close()

	client := httputil.NewClientConn(c, nil)
	defer client.Close()

	image := in.imageId
	digest := in.digest
	log.Printf("Processing digest: %s\n", digest)
	// save image
	img_dir_name := strings.Replace(image, ":", "_", -1)
	img_dir_name = strings.Replace(img_dir_name, "/", "_", -1)
	path := fmt.Sprintf("%s/%s", scannerOutputDir, img_dir_name)
	log.Println(path)

	if strings.Contains(path, "<none>") {
		log.Printf("WARNING: Image : %s won't be scanned.", digest)
	} else {
		newTarPath, err := saveImageToTar(client, digest, path)

		if err != nil {
			log.Printf("Error while making tar file: %s\n", err)
		} else {
			scanImage(newTarPath, image, digest)
		}

		if !*dumpFlag {
			os.RemoveAll(path)
		}
	}

	log.Println("Finished")
}
