package configutil

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

func ReadConfig(filePath string, config interface{}) error {

	if len(filePath) > 0 {
		err := readCommandYaml(filePath, config)
		if err != nil {
			return fmt.Errorf("Unable to load Yaml file '%s': inner error: \n'%v'", filePath, err.Error())
		}
	} else {
		err := readYamlFromStdin(config)
		if err != nil {
			return fmt.Errorf("Unalbe to read Yaml from stdin: \n'%v'", err.Error())
		}
	}

	return nil
}

func readCommandYaml(filePath string, container interface{}) (err error) {
	b, err := loadDataFrom(filePath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(b, container)
	if err != nil {
		return err
	}

	err = nil
	return
}

func readYamlFromStdin(container interface{}) (err error) {
	b, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(b, container)
	if err != nil {
		return err
	}

	err = nil
	return
}

func loadDataFrom(filePath string) ([]byte, error) {
	u, err := url.Parse(filePath)
	if err != nil {
		return nil, err
	}

	if u.Host != "" { // i have a host, now do i support the scheme?
		switch u.Scheme {
		case "http", "https":
			res, err := http.Get(filePath)
			if err != nil {
				return nil, err
			}
			return ioutil.ReadAll(res.Body)
		default:
			return nil, fmt.Errorf("scheme of %s is unsupported", filePath)
		}
	} else if u.Path != "" { // i dont have a host, but I have a path. I am a local file.
		if _, notFoundFileErr := os.Stat(filePath); notFoundFileErr != nil {
			return nil, fmt.Errorf("Cannot read from file: '%s' because it does not exist.", filePath)
		}
		return ioutil.ReadFile(filePath)
	} else if runtime.GOOS == "windows" && strings.Contains(u.String(), "\\") {
		// on Windows systems u.Path is always empty, so we need to check the string directly.
		if _, notFoundFileErr := os.Stat(filePath); notFoundFileErr != nil {
			return nil, fmt.Errorf("Cannot read from file: '%s' because it does not exist.", filePath)
		}
		return ioutil.ReadFile(filePath)
	} else {
		return nil, fmt.Errorf("unable to determine how to load from path %s", filePath)
	}
}
