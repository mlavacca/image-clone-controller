package controllers

import (
	"image-clone-controller/pkg/imagesManagement"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"time"
)

const (
	systemNamespace = "kube-system"
	requeuePeriod   = 10 * time.Second
)

func containerIterator(containers []v1.Container) (bool, error) {
	var toPatch bool
	for i, c := range containers {
		backupImageName, err := imagesManagement.Get().EnforceBackup(c.Image)
		if err != nil {
			return false, err
		}
		if backupImageName != c.Image {
			containers[i].Image = backupImageName
			toPatch = true
		}
	}

	return toPatch, nil
}

func commonPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(createEvent event.CreateEvent) bool {
			return createEvent.Object.GetNamespace() != systemNamespace
		},
		DeleteFunc: func(deleteEvent event.DeleteEvent) bool {
			return false
		},
		UpdateFunc: func(updateEvent event.UpdateEvent) bool {
			return updateEvent.ObjectNew.GetNamespace() != systemNamespace
		},
		GenericFunc: func(genericEvent event.GenericEvent) bool {
			return genericEvent.Object.GetNamespace() != systemNamespace
		},
	}
}
