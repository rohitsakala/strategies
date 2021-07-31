package browser

import (
	"github.com/tebeka/selenium"
)

type ChromeBrowser struct {
	BinaryPath string
	Service    *selenium.Service
	Port       int
}

func NewChromeBrowser(binaryPath string, port int) ChromeBrowser {
	return ChromeBrowser{
		BinaryPath: binaryPath,
		Port:       port,
	}
}

func (k ChromeBrowser) Start() error {
	var err error

	k.Service, err = selenium.NewChromeDriverService("/usr/local/bin/chromedriver", k.Port)
	if err != nil {
		return err
	}

	return nil
}

func (k ChromeBrowser) Stop() error {
	err := k.Service.Stop()
	if err != nil {
		return err
	}

	return nil
}
