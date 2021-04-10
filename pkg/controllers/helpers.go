package controllers

import (
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"time"
)

const (
	systemNamespace = "kube-system"
	requeuePeriod = 10 * time.Second
)

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
