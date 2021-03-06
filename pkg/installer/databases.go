package installer

import (
	"bytes"

	"github.com/pkg/errors"
	"github.com/schemahero/schemahero/pkg/client/schemaheroclientset/scheme"
	extensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	extensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	extensionsscheme "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/scheme"
	extensionsv1client "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	extensionsv1beta1client "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	kuberneteserrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/rest"
)

func databasesCRDYAML(useExtensionsv1beta1 bool) ([]byte, error) {
	s := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme.Scheme, scheme.Scheme)
	var result bytes.Buffer
	if useExtensionsv1beta1 {
		if err := s.Encode(databasesCRDV1Beta1(), &result); err != nil {
			return nil, errors.Wrap(err, "failed to marshal databases v1beta1 crd")
		}

	} else {
		if err := s.Encode(databasesCRDV1(), &result); err != nil {
			return nil, errors.Wrap(err, "failed to marshal databases v1 crd")
		}

	}

	return result.Bytes(), nil
}

func ensureDatabasesCRD(cfg *rest.Config, useExtensionsv1beta1 bool) error {
	if useExtensionsv1beta1 {
		extensionsClient, err := extensionsv1beta1client.NewForConfig(cfg)
		if err != nil {
			return errors.Wrap(err, "faild to create extensions client")
		}

		_, err = extensionsClient.CustomResourceDefinitions().Get("databases.databases.schemahero.io", metav1.GetOptions{})
		if err != nil {
			if !kuberneteserrors.IsNotFound(err) {
				return errors.Wrap(err, "failed to get databases crd")
			}

			_, err = extensionsClient.CustomResourceDefinitions().Create(databasesCRDV1Beta1())
			if err != nil {
				return errors.Wrap(err, "failed to create databases crd")
			}
		}

		return nil
	}

	extensionsClient, err := extensionsv1client.NewForConfig(cfg)
	if err != nil {
		return errors.Wrap(err, "faild to create extensions client")
	}

	_, err = extensionsClient.CustomResourceDefinitions().Get("databases.databases.schemahero.io", metav1.GetOptions{})
	if err != nil {
		if !kuberneteserrors.IsNotFound(err) {
			return errors.Wrap(err, "failed to get databases crd")
		}

		_, err := extensionsClient.CustomResourceDefinitions().Create(databasesCRDV1())
		if err != nil {
			return errors.Wrap(err, "failed to create databases crd")
		}
	}

	return nil
}

func databasesCRDV1() *extensionsv1.CustomResourceDefinition {
	extensionsscheme.AddToScheme(scheme.Scheme)
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode([]byte(generatedDatabaseCRDV1), nil, nil)
	if err != nil {
		panic(err) // todo
	}

	return obj.(*extensionsv1.CustomResourceDefinition)
}

func databasesCRDV1Beta1() *extensionsv1beta1.CustomResourceDefinition {
	extensionsscheme.AddToScheme(scheme.Scheme)
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode([]byte(generatedDatabaseCRDV1Beta1), nil, nil)
	if err != nil {
		panic(err) // todo
	}

	return obj.(*extensionsv1beta1.CustomResourceDefinition)
}
