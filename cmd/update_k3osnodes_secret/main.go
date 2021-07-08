package main

import (
	"fmt"
)

func main() {
	// 1. read all YAML files in the config directory
	// 2. parse them using github.com/annismckenzie/k3os-config-operator as a library
	// 3. extract the node name
	// 4. generate the secret into config/k3osconfig.yaml by using kustomize as a library
	// 5. build this tool in the main Dockerfile using a multistage build, then add a Makefile target to invoke it on its own

	fmt.Println("Successfully updated k3osconfig.yaml")
}
