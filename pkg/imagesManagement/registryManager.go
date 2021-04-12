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
	backupRepository string
}

var registryConnectionInstance *registryConnection

func SetupRegistryManager(backupRegistry, backupRepository string) error {
	if registryConnectionInstance != nil {
		return errors.New("registryManager already initialized")
	}

	registryConnectionInstance = &registryConnection{
		backupRegistry: backupRegistry,
		backupRepository: backupRepository,
	}
	return nil
}

func Get() RegistryManager {
	return registryConnectionInstance
}

func (r *registryConnection) EnforceBackup(originalImageName string) (string, error) {
	if r == nil {
		return "", errors.New("registryManager not initialized")
	}

	// get the original image's reference
	originalRef, err := name.ParseReference(originalImageName)
	if err != nil {
		return "", err
	}

	// check if the image already belongs to the backup repository
	if originalRef.Context().Registry.RegistryStr() == r.backupRegistry &&
		strings.HasPrefix(originalRef.Context().RepositoryStr(), fmt.Sprintf("%s/", r.backupRepository)){
		return originalImageName, nil
	}

	// build the backup image reference
	backupImageName := fmt.Sprintf("%s/%s:%s",
		r.backupRepository,
		strings.Replace(originalRef.Context().RepositoryStr(), "/", "-", -1),
		originalRef.Identifier())

	backupRef, err := name.ParseReference(backupImageName)
	if err != nil {
		return "", err
	}

	// pull the backup image
	klog.V(3).Infof("pulling the backup image %s", backupRef.String())
	_, err = remote.Image(backupRef, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err == nil {
		klog.V(3).Info("backup image %s already existing", backupRef.String())
		return backupRef.String(), nil
	}
	if !isImageNotFound(err) {
		return "", err
	}

	// pull the original image
	klog.Infof("backup image %s does not exist", backupRef.String())
	klog.V(3).Infof("pulling the original image %s", originalRef.String())
	originalImage, err := remote.Image(originalRef, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return "", err
	}

	// if the backup image does not exist yet, create and push it
	klog.V(3).Info("pushing the backup image %s", backupRef.String())
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
