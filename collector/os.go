package collector

import (
	"os/exec"
	"strings"
)

type Os struct {
	Distribution  string `json:"distribution"`
	KernelVersion string `json:"kernel"`
}

func GetOs() Os {
	c := Os{}
	c.getDistro()
	c.getKernel()
	return c
}

func (c *Os) getDistro() {
	// Linux
	if lsb, ok := which("lsb_release"); ok {
		std, err := exec.Command(lsb, "-d", "-s").CombinedOutput()
		if err == nil {
			c.Distribution = trim(std)
			return
		}
	}

	// OS X
	if swvers, ok := which("sw_vers"); ok {
		std, err := exec.Command(swvers, "-productVersion").CombinedOutput()
		if err == nil {
			c.Distribution = "Mac OS X " + trim(std)
			return
		}
	}

	c.Distribution = "Unknown"
}

func (c *Os) getKernel() {
	std, err := exec.Command("uname", "-s", "-r").CombinedOutput()
	if err == nil {
		c.KernelVersion = trim(std)
		return
	}

	c.KernelVersion = "Unknown"
}

func trim(b []byte) string {
	return strings.TrimRight(string(b), "\r\n\t ")
}

func which(cmd string) (string, bool) {
	std, err := exec.Command("which", cmd).CombinedOutput()
	if err == nil {
		return trim(std), true
	}

	return "", false
}
