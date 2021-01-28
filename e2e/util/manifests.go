package util

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os/exec"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	serializeryaml "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

// ApplyCRD applies custom resource definitions for kms-secrets which is located in config/crd.
func ApplyCRD(ctx context.Context, cfg *rest.Config) error {
	buf, err := exec.Command("kustomize", "build", "../config/crd").Output()
	if err != nil {
		return err
	}
	return apply(ctx, cfg, buf)
}

// DeleteCRD deletes custom resource definitions for kms-secrets.
func DeleteCRD(ctx context.Context, cfg *rest.Config) error {
	buf, err := exec.Command("kustomize", "build", "../config/crd").Output()
	if err != nil {
		return err
	}
	return delete(ctx, cfg, buf)
}

// ApplyRBAC applies role based access control for operator which is located in config/rbac.
func ApplyRBAC(ctx context.Context, cfg *rest.Config) error {
	buf, err := exec.Command("kustomize", "build", "../config/rbac").Output()
	if err != nil {
		return err
	}
	return apply(ctx, cfg, buf)
}

// DeleteRBAC deletes role based access control for operator.
func DeleteRBAC(ctx context.Context, cfg *rest.Config) error {
	buf, err := exec.Command("kustomize", "build", "../config/rbac").Output()
	if err != nil {
		return err
	}
	return delete(ctx, cfg, buf)
}

func apply(ctx context.Context, cfg *rest.Config, data []byte) error {
	return restAction(ctx, cfg, data, func(ctx context.Context, dr dynamic.ResourceInterface, obj *unstructured.Unstructured, data []byte) error {
		klog.Infof("applying: %v", obj.GetName())
		_, err := dr.Patch(ctx, obj.GetName(), types.ApplyPatchType, data, metav1.PatchOptions{
			FieldManager: "e2e",
		})
		if err != nil {
			klog.Error(err)
		}
		return err
	})
}

func delete(ctx context.Context, cfg *rest.Config, data []byte) error {
	return restAction(ctx, cfg, data, func(ctx context.Context, dr dynamic.ResourceInterface, obj *unstructured.Unstructured, data []byte) error {
		return dr.Delete(ctx, obj.GetName(), metav1.DeleteOptions{})
	})
}

type actionFunc func(ctx context.Context, dr dynamic.ResourceInterface, obj *unstructured.Unstructured, data []byte) error

func restAction(ctx context.Context, cfg *rest.Config, data []byte, action actionFunc) error {
	dc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return err
	}

	multidocReader := utilyaml.NewYAMLReader(bufio.NewReader(bytes.NewReader(data)))

	for {
		buf, err := multidocReader.Read()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		var typeMeta runtime.TypeMeta
		if err := yaml.Unmarshal(buf, &typeMeta); err != nil {
			continue
		}
		if typeMeta.Kind == "" {
			continue
		}

		obj := &unstructured.Unstructured{}
		_, gvk, err := serializeryaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(buf, nil, obj)
		if err != nil {
			return err
		}

		mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			return err
		}

		var dr dynamic.ResourceInterface
		if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
			dr = dyn.Resource(mapping.Resource).Namespace(obj.GetNamespace())
		} else {
			dr = dyn.Resource(mapping.Resource)
		}

		// Remove status field before apply
		unstructured.RemoveNestedField(obj.Object, "status")

		data, err := json.Marshal(obj)
		if err != nil {
			return err
		}

		if err := action(ctx, dr, obj, data); err != nil {
			return err
		}

	}
}
