package cfg

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/beevik/guid"
)

type Settings struct {
	Host                    string
	ProjectName             string
	IntervalTimeInSecond    time.Duration
	FirstTimeLoadRetryCount int
}

type configuration struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type configurationDto struct {
	key   string
	value string
}

func (c configurationDto) String() string {
	return c.value
}

func (c configurationDto) Int() (int, error) {
	n, err := strconv.Atoi(c.value)

	if err != nil {
		return 0, err
	}

	return n, nil
}

type Environments struct {
	FilePath string
	FileName string
}

var client = &http.Client{
	Timeout: time.Second * 30,
}

var configurationDtoList = make(map[string]configurationDto)
var interval = make(chan bool, 1)
var Complete = make(chan bool, 1)
var defaultRequestDelayInSecond = time.Second * 5

func Get(configurationKey string) configurationDto {
	return configurationDtoList[configurationKey]
}

func (s Settings) getConfigurationsFromService() ([]configuration, error) {
	req, err := http.NewRequest("GET", s.Host+"/configurations?projectName="+s.ProjectName, nil)

	if err != nil {
		log.Println("an error occurred while getting configurations.", err)
	}

	req.Header.Add("x-correlationId", guid.New().String())
	req.Header.Add("x-agentName", "cfg-go-client")

	resp, err := client.Do(req)
	var configurationList []configuration

	if resp != nil {
		err = json.NewDecoder(resp.Body).Decode(&configurationList)

		if err != nil {
			log.Println("an error occurred while getting configurations...", err)
		}

		defer resp.Body.Close()

		return configurationList, nil
	}

	return configurationList, errors.New("an error occurred")
}

type loader interface {
	loadConfigurationsFromService(settings Settings) error
}

func (s Settings) loadConfigurationsFromService() error {
	f := true
	init := false
	counter := 0

	for {
		<-interval
		configurationList, err := s.getConfigurationsFromService()

		if err != nil {
			log.Printf("configurations didn't updated counter:%v", counter)

			if counter < s.FirstTimeLoadRetryCount && !init {
				time.Sleep(defaultRequestDelayInSecond)
				counter++
				interval <- true
				continue
			}

			if init {
				time.Sleep(s.IntervalTimeInSecond * time.Second)
				interval <- true
				continue
			}

			log.Println("arrived max retry count before first load")
			os.Exit(0)

		}

		for _, config := range configurationList {
			if cDto, ok := configurationDtoList[config.Key]; ok {
				cDto.value = config.Value
			} else {
				log.Printf("new configuration found. configurationKey: %v configuration new value: %v", config.Key, config.Value)
				configurationDtoList[config.Key] = configurationDto{
					key:   config.Key,
					value: config.Value,
				}
			}
		}

		log.Println("configurations are reloaded")

		if f {
			Complete <- true
			init = true
			close(Complete)
			f = false
		}

		time.Sleep(s.IntervalTimeInSecond * time.Second)
		interval <- true
		continue
	}
}

func Init(settings Settings) error {
	if len(settings.Host) == 0 || len(settings.ProjectName) == 0 {
		log.Println("Settings must be valid")
		return errors.New("settings must be valid")
	}

	interval <- true
	go settings.loadConfigurationsFromService()

	return nil
}
