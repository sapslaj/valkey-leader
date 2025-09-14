package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/sapslaj/valkey-leader/yoke/v1/crd"
	"github.com/sapslaj/valkey-leader/yoke/v1/valkey"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	var v crd.Valkey
	err := yaml.NewYAMLToJSONDecoder(os.Stdin).Decode(&v)
	if err != nil && err != io.EOF {
		return err
	}

	return json.NewEncoder(os.Stdout).Encode(valkey.Create(v))
}
