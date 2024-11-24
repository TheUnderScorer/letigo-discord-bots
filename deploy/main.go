package main

import (
	"app/server/responses"
	"encoding/json"
	"github.com/charmbracelet/log"
	"net/http"
	"os"
	"os/exec"
)

const service = "prod"

func main() {
	currentVersion := getCurrentVersion()
	runningVersion := getRunningVersion()

	log.Infof("current version: %s", currentVersion)
	log.Infof("running version: %s", runningVersion)

	if currentVersion == runningVersion {
		log.Info("version is up to date")
		return
	}

	stopService()
	runDockerComposeCommand()
}

func getCurrentVersion() string {
	packageJson := struct {
		Version string `json:"version"`
	}{}

	file, err := os.Open("../package.json")
	if err != nil {
		log.Fatalf("failed to open package.json: %s", err)
	}

	err = json.NewDecoder(file).Decode(&packageJson)
	if err != nil {
		log.Fatalf("failed to decode package.json: %s", err)
	}

	return packageJson.Version
}

func getRunningVersion() string {
	var result responses.VersionInfo
	response, err := http.Get("http://localhost:3000")
	if err != nil {
		log.Fatalf("failed to get version: %s", err)
	}

	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		log.Fatalf("failed to decode version: %s", err)
	}

	return result.Version
}

func stopService() {
	cmd := exec.Command("docker", "compose", "down")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		log.Fatalf("failed to stop service: %s", err)
	}
}

func runDockerComposeCommand() {
	cmd := exec.Command("docker", "compose", "--build", "up", "-d", service)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		log.Fatalf("failed to run docker compose command: %s", err)
	}
}
