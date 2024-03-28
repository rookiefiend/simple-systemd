package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/ghodss/yaml"
	"github.com/sirupsen/logrus"
)

type Service struct {
	Name        string            `json:"name,omitempty"`
	Description string            `json:"description,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
	Cwd         string            `json:"cwd,omitempty"`
	ExecStart   []string          `json:"execStart,omitempty"`
	RestartSec  int               `json:"restartSec,omitempty"`
	Depends     []string          `json:"depends,omitempty"`
	Status      bool              `json:"status,omitempty"`
}

var (
	configDir  = flag.String("c", "./config", "set service config file scan dir")
	serviceMap = make(map[string]*Service, 5)
)

func main() {
	flag.Parse()
	err := ScanServcie()
	if err != nil {
		logrus.WithField("config", *configDir).
			WithError(err).Fatalf("failed to scan service")
	}
	for _, svr := range serviceMap {
		go ManageService(svr)
	}
	// 创建一个信号接收通道
	sigCh := make(chan os.Signal, 1)

	// 监听指定的系统信号
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}

func ScanServcie() error {
	err := filepath.Walk(*configDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			logrus.WithError(err).Warnf("failed to walk through")
		}
		if info.IsDir() {
			return err
		}
		if !strings.HasSuffix(info.Name(), "service.yaml") {
			return nil
		}
		fContent, err := os.ReadFile(path)
		if err != nil {
			logrus.WithField("path", path).
				WithError(err).Warnf("failed to open path")
			return err
		}
		svr := Service{}
		err = yaml.Unmarshal(fContent, &svr)
		if err != nil {
			logrus.WithField("path", path).
				WithField("content", string(fContent)).
				WithError(err).Error("failed to decode config file")
			return err
		}
		svr.Status = false
		_, ok := serviceMap[svr.Name]
		if ok {
			logrus.Fatalf("service [%s] is duplicated", svr.Name)
		}
		if svr.RestartSec == 0 {
			svr.RestartSec = 5
		}
		serviceMap[svr.Name] = &svr
		return nil
	})
	if err != nil {
		logrus.WithError(err).Errorf("failed to scan config file")
		return err
	}
	return nil
}

func ManageService(svr *Service) {
	for {
		for _, dep := range svr.Depends {
			for {
				svr, ok := serviceMap[dep]
				if !ok {
					logrus.WithField("dep", dep).Errorf("depend is not found")
				} else {
					if svr.Status {
						break
					}
				}
				time.Sleep(1 * time.Second)
			}
		}

		cmd := exec.Command(svr.ExecStart[0], svr.ExecStart[1:]...)
		cmd.Env = os.Environ()
		if svr.Cwd != "" {
			cmd.Dir = svr.Cwd
		}
		cmd.Stdout = NewServiceWrapWriter(svr.Name, os.Stdout)
		cmd.Stderr = NewServiceWrapWriter(svr.Name, os.Stdout)
		for k, v := range svr.Env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
		timer := time.NewTimer(1 * time.Second)
		go func() {
			_, isOpen := <-timer.C
			if isOpen {
				svr.Status = true
			} else {
				svr.Status = false
			}
		}()
		err := cmd.Run()
		if err != nil {
			logrus.WithField("command", cmd.String()).
				WithError(err).
				Warnf("failed to run command")
		}
		timer.Stop()
		time.Sleep(time.Duration(svr.RestartSec) * time.Second)
	}
}

type ServcieWrapWriter struct {
	name string
	f    *os.File
}

func NewServiceWrapWriter(name string, f *os.File) *ServcieWrapWriter {
	return &ServcieWrapWriter{
		name: name,
		f:    f,
	}
}

func (w ServcieWrapWriter) Write(p []byte) (n int, err error) {
	prefix := []byte(fmt.Sprintf("[%s]", w.name))
	p = append(prefix, p...)
	n, err = w.f.Write(p)
	if n > len(prefix) {
		n -= len(prefix)
	} else {
		n = 0
	}
	return n, err
}
