package config

import (
    "gopkg.in/yaml.v2"
    "io/ioutil"
    "os"
)

type HostConfig struct {
    Hosts []Host `yaml:"hosts"`
}

type Host struct {
    Name     string `yaml:"name"`
    Hostname string `yaml:"hostname"`
    Username string `yaml:"user"`
    Password string `yaml:"password"`
}

func LoadConfig(filePath string) (*HostConfig, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    data, err := ioutil.ReadAll(file)
    if err != nil {
        return nil, err
    }

    var config HostConfig
    err = yaml.Unmarshal(data, &config)
    if err != nil {
        return nil, err
    }

    return &config, nil
}