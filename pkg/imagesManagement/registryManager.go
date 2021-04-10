package imagesManagement

import (
	"errors"
	"fmt"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"k8s.io/klog/v2"
	"net/http"
	"strings"
)

type RegistryManager interface {
	EnforceBackup(repository string) (string, error)
}

type registryConnection struct {
	backupRegistry string
}

var registryConnectionInstance *registryConnection

func SetupRegistryManager(backupRegistry string) error {
	if registryConnectionInstance != nil {
		return errors.New("registryManager already setup")
	}

	registryConnectionInstance = &registryConnection{
		backupRegistry: backupRegistry,
	}
	return nil
}

func Get() RegistryManager {
	return registryConnectionInstance
}

func (r *registryConnection) EnforceBackup(originalImageName string) (string, error) {
	if strings.HasPrefix(fmt.Sprintf("%s/", originalImageName), r.backupRegistry) {
		return originalImageName, nil
	}

	// get the original image's reference
	originalRef, err := name.ParseReference(originalImageName)
	if err != nil {
		return "", err
	}

	// build the backup image reference
	backupImageName := fmt.Sprintf("%s/%s:%s",
		r.backupRegistry,
		strings.Replace(originalRef.Context().RepositoryStr(), "/", "-", -1),
		originalRef.Identifier())

	if backupImageName == originalImageName {
		return originalImageName, nil
	}

	backupRef, err := name.ParseReference(backupImageName)
	if err != nil {
		return "", err
	}

	// pull the original image
	originalImage, err := remote.Image(originalRef, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return "", err
	}

	// pull the backup image
	_, err = remote.Image(backupRef, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err == nil {
		return backupRef.String(), nil
	}
	if !isImageNotFound(err) {
		return "", err
	}

	// if the backup image does not exist yet, create and push it
	if err = pushBackupImage(originalImage, backupRef.String()); err != nil {
		return "", err
	}

	klog.Infof("backup of image %s to %s completed", originalRef, backupRef)

	return backupRef.String(), nil
}

func pushBackupImage(originalImage v1.Image, backupName string) error {
	backupImg, err := mutate.AppendLayers(originalImage)
	if err != nil {
		return err
	}

	if err = crane.Push(backupImg, backupName); err != nil {
		return err
	}

	return nil
}

func isImageNotFound(err error) bool {
	return err.(*transport.Error).StatusCode == http.StatusNotFound
}
