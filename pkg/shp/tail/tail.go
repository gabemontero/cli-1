package tail

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// Tail represents a "tail" command streaming log outputs to stdout interface, and errors are written
// to stderr interface directly.
type Tail struct {
	ctx       context.Context      // global context
	clientset kubernetes.Interface // kubernetes client instance
	stopCh    chan bool            // stop channel

	stdout io.Writer
	stderr io.Writer
}

// SetStdout set and alternative stdout writer.
func (t *Tail) SetStdout(w io.Writer) {
	t.stdout = w
}

// SetStderr set and alternative stderr writer.
func (t *Tail) SetStderr(w io.Writer) {
	t.stderr = w
}

// Start start streaming logs for informed target.
func (t *Tail) Start(ns, podName, container string) {
	go func() {
		podClient := t.clientset.CoreV1().Pods(ns)
		stream, err := podClient.GetLogs(podName, &corev1.PodLogOptions{
			Follow:    true,
			Container: container,
		}).Stream(t.ctx)
		if err != nil {
			fmt.Fprintln(t.stderr, err)
			return
		}
		defer stream.Close()

		go func() {
			<-t.stopCh
			stream.Close()
		}()

		containerName := strings.TrimPrefix(container, "step-")
		sc := bufio.NewScanner(stream)
		for sc.Scan() {
			fmt.Fprintf(t.stdout, "[%s] %s\n", containerName, sc.Text())
		}
	}()
	go func() {
		<-t.ctx.Done()
		close(t.stopCh)
	}()
}

// Stop closes stop channel to stop log streaming.
func (t *Tail) Stop() {
	close(t.stopCh)
}

// NewTail instantiate Tail, using by default regular stdout and stderr.
func NewTail(ctx context.Context, clientset kubernetes.Interface) *Tail {
	return &Tail{
		ctx:       ctx,
		clientset: clientset,
		stopCh:    make(chan bool, 1),
		stdout:    os.Stdout,
		stderr:    os.Stderr,
	}
}
