package reactor

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

// PodWatcher a simple function orchestrator based on watching a given pod and reacting upon the
// state modifications, should work as a helper to build business logic based on the build POD
// changes.
type PodWatcher struct {
	ctx     context.Context
	stopCh  chan bool       // stops the event loop execution
	watcher watch.Interface // client watch instance

	skipPodFn       SkipPodFn
	onPodAddedFn    OnPodEventFn
	onPodModifiedFn OnPodEventFn
	onPodDeletedFn  OnPodEventFn
}

// SkipPodFn a given pod instance is informed and expects a boolean as return. When true is returned
// this container state processing is skipped completely.
type SkipPodFn func(pod *corev1.Pod) bool

// OnPodEventFn when a pod is modified this method handles the event.
type OnPodEventFn func(pod *corev1.Pod) error

// WithSkipPodFn sets the skip function instance.
func (p *PodWatcher) WithSkipPodFn(fn SkipPodFn) *PodWatcher {
	p.skipPodFn = fn
	return p
}

// WithOnPodAddedFn sets the function executed when a pod is added.
func (p *PodWatcher) WithOnPodAddedFn(fn OnPodEventFn) *PodWatcher {
	p.onPodAddedFn = fn
	return p
}

// WithOnPodModifiedFn sets the funcion executed when a pod is modified.
func (p *PodWatcher) WithOnPodModifiedFn(fn OnPodEventFn) *PodWatcher {
	p.onPodModifiedFn = fn
	return p
}

// WithOnPodDeletedFn sets the funcion executed when a pod is modified.
func (p *PodWatcher) WithOnPodDeletedFn(fn OnPodEventFn) *PodWatcher {
	p.onPodDeletedFn = fn
	return p
}

// Start runs the event loop based on a watch instantiated against informed pod. In case of errors
// the loop is interrupted.
func (p *PodWatcher) Start() (*corev1.Pod, error) {
	for {
		select {
		// handling the regular pod modification events, which should trigger calling event functions
		// accordinly
		case event := <-p.watcher.ResultChan():
			if event.Object == nil {
				continue
			}
			pod, ok := event.Object.(*corev1.Pod)
			if !ok {
				continue
			}

			if p.skipPodFn != nil && p.skipPodFn(pod) {
				continue
			}

			switch event.Type {
			case watch.Added:
				if p.onPodAddedFn != nil {
					if err := p.onPodAddedFn(pod); err != nil {
						return pod, err
					}
				}
			case watch.Modified:
				if p.onPodModifiedFn != nil {
					if err := p.onPodModifiedFn(pod); err != nil {
						return pod, err
					}
				}
			case watch.Deleted:
				if p.onPodDeletedFn != nil {
					if err := p.onPodDeletedFn(pod); err != nil {
						return pod, err
					}
				}
			}
		// watching over global context, when done is informed on the context it needs to reflect on
		// the event loop as well.
		case <-p.ctx.Done():
			p.watcher.Stop()
			return nil, nil
		// watching over stop channel to stop the event loop on demand.
		case <-p.stopCh:
			p.watcher.Stop()
			return nil, nil
		}
	}
}

// Stop closes the stop channel, and stops the execution loop.
func (p *PodWatcher) Stop() {
	close(p.stopCh)
}

// NewPodWatcher instantiate PodWatcher event-loop.
func NewPodWatcher(
	ctx context.Context,
	clientset kubernetes.Interface,
	listOpts metav1.ListOptions,
	ns string,
) (*PodWatcher, error) {
	w, err := clientset.CoreV1().Pods(ns).Watch(ctx, listOpts)
	if err != nil {
		return nil, err
	}
	return &PodWatcher{ctx: ctx, watcher: w, stopCh: make(chan bool)}, nil
}
