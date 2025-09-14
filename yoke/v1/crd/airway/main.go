package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/yokecd/yoke/pkg/apis/airway/v1alpha1"
	"github.com/yokecd/yoke/pkg/openapi"

	"github.com/sapslaj/valkey-leader/yoke/v1/crd"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	return json.NewEncoder(os.Stdout).Encode(v1alpha1.Airway{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("backends.%s", crd.APIGroup),
		},
		Spec: v1alpha1.AirwaySpec{
			Mode: v1alpha1.AirwayModeStandard,
			WasmURLs: v1alpha1.WasmURLs{
				Flight: "https://github.com/sapslaj/valkey-leader/releases/download/latest/valkey-leader-yoke-flight.wasm.gz",
			},
			Template: apiextv1.CustomResourceDefinitionSpec{
				Group: crd.APIGroup,
				Names: apiextv1.CustomResourceDefinitionNames{
					Plural:   "valkeys",
					Singular: "valkey",
					Kind:     crd.KindValkey,
				},
				Scope: apiextv1.NamespaceScoped,
				Versions: []apiextv1.CustomResourceDefinitionVersion{
					{
						Name:    "v1",
						Served:  true,
						Storage: true,
						Schema: &apiextv1.CustomResourceValidation{
							OpenAPIV3Schema: openapi.SchemaFrom(reflect.TypeFor[crd.Valkey]()),
						},
					},
				},
			},
		},
	})
}
