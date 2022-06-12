package serviceentry

import (
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	"k8s.io/client-go/tools/cache"
)

// NewHandler returns an operator-sdk Handler which updates the store based on Kubernetes events
func AttachHandler(model ServiceEntryModel, informer cache.SharedIndexInformer) {
	informer.AddEventHandler(handler{model})
}

// Implements operator-sdk.Handler; we use it to update our representation of service entries.
type handler struct {
	ServiceEntryModel
}

func (c handler) OnAdd(obj interface{}) {
	// TODO: handle the case this isn't wrong and log, bail out w/o calling insert
	se := obj.(*v1alpha3.ServiceEntry)
	c.Insert(se)
}

func (c handler) OnUpdate(oldObj, newObj interface{}) {
	old := oldObj.(*v1alpha3.ServiceEntry)
	se := newObj.(*v1alpha3.ServiceEntry)
	c.Update(old, se)
}

func (c handler) OnDelete(obj interface{}) {
	se := obj.(*v1alpha3.ServiceEntry)
	c.Delete(se)
}
