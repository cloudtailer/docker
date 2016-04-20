package daemon

import (
	"fmt"
	"strings"
	"time"
)

const (
	statusRunning = "running"
	statusExited  = "exited"
	statusRemoved = "removed"
)

var (
	supportedWaitStatus = map[string]struct{}{
		statusRunning: {},
		statusExited:  {},
		statusRemoved: {},
	}
)

// ContainerWait stops processing until the given container is
// stopped. If the container is not found, an error is returned. On a
// successful stop, the exit code of the container is returned. On a
// timeout, an error is returned. If you want to wait forever, supply
// a negative duration for the timeout.
func (daemon *Daemon) ContainerWait(name string, timeout time.Duration, custom string) (int, error) {
	custom = strings.TrimSpace(custom)
	if len(custom) == 0 {
		container, err := daemon.GetContainer(name)
		if err != nil {
			return -1, err
		}

		return container.WaitStop(timeout)
	}

	status, err := daemon.validateWaitCustomSeq(custom)
	if err != nil {
		return -1, err
	}

	pidOrEcode := 0
	for _, s := range status {
		switch s {
		case statusRunning:
			if pidOrEcode, err = daemon.waitRunning(name, timeout); err != nil {
				return pidOrEcode, err
			}
		case statusExited:
			if pidOrEcode, err = daemon.waitExited(name, timeout); err != nil {
				return pidOrEcode, err
			}
		case statusRemoved:
			if err = daemon.waitRemoved(name, timeout); err != nil {
				return pidOrEcode, err
			}
		}
	}

	return pidOrEcode, nil
}

func (daemon *Daemon) validateWaitCustomSeq(custom string) ([]string, error) {
	if len(custom) == 0 {
		return nil, fmt.Errorf("custom string can't be empty")
	}
	status := strings.Split(custom, ",")
	for _, s := range status {
		if _, ok := supportedWaitStatus[s]; !ok {
			return nil, fmt.Errorf("Unsupported custom string %q", s)
		}
	}

	return status, nil
}

func (daemon *Daemon) waitRunning(name string, timeout time.Duration) (int, error) {
	container, err := daemon.GetContainer(name)
	if err != nil {
		return -1, err
	}

	return container.WaitRunning(timeout)
}

func (daemon *Daemon) waitExited(name string, timeout time.Duration) (int, error) {
	container, err := daemon.GetContainer(name)
	if err != nil {
		return -1, err
	}

	return container.WaitStop(timeout)
}

func (daemon *Daemon) waitRemoved(name string, timeout time.Duration) error {
	var alarm <-chan time.Time
	if timeout > 0 {
		alarm = time.After(timeout)
	}
	for {
		container, _ := daemon.GetContainer(name)
		if container == nil { // FIXME: need check err also
			break
		}

		if timeout > 0 {
			select {
			case <-alarm:
				return fmt.Errorf("wait container to be removed timeout")
			default:
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	return nil
}
