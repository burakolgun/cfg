package cfg

import (
	"encoding/json"
	"errors"
	"fmt"
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

type ConfigurationDto struct {
	key   string
	value string
}

func (c ConfigurationDto) String() string {
	return c.value
}

func (c ConfigurationDto) Int() int {
	n, err := strconv.Atoi(c.value)

	if err != nil {
		panic(err)
	}

	return n
}

func (c ConfigurationDto) Bool() bool {
	n, err := strconv.ParseBool(c.value)

	if err != nil {
		panic(err)
	}

	return n
}

func GetEnvironmentVariable(v string) ConfigurationDto {
	return ConfigurationDto{
		key:   v,
		value: os.Getenv(v),
	}
}

type Environments struct {
	FilePath string
	FileName string
}

var client = &http.Client{
	Timeout: time.Second * 30,
}

var configurationDtoList = make(map[string]ConfigurationDto)
var interval = make(chan bool, 1)
var Complete = make(chan bool, 1)
var defaultRequestDelayInSecond = time.Second * 5

func Get(configurationKey string) ConfigurationDto {
	return configurationDtoList[configurationKey]
}

func (s Settings) getConfigurationsFromService() ([]configuration, error) {
	req, err := http.NewRequest("GET", s.Host+"/configurations?projectName="+s.ProjectName, nil)

	if err != nil {
		log.Println("an error occurred while getting configurations.", err)
	}

	req.Header.Add("x-correlationid", guid.New().String())
	req.Header.Add("x-agentname", "cfg-go-client")

	resp, err := client.Do(req)
	fmt.Println("requ")
	var configurationList []configuration

	if resp != nil {
		err = json.NewDecoder(resp.Body).Decode(&configurationList)

		if err != nil {
			log.Println("an error occurred while getting configurations...", err)
			return []configuration{}, errors.New("an error occurred while decoding response")
		}

		defer resp.Body.Close()

		return configurationList, nil
	}

	return configurationList, errors.New("an error occurred")
}

type loader interface {
	loadConfigurationsFromService(settings Settings)
}

func (s Settings) loadConfigurationsFromService() {
	f := true
	init := false
	counter := 0

	for {
		<-interval
		configurationList, err := s.getConfigurationsFromService()

		if err != nil && f {
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

			panic("arrived max retry count before first load")
		}

		for _, config := range configurationList {
			if cDto, ok := configurationDtoList[config.Key]; ok {
				if cDto.value != config.Value {
					if !f {
						log.Printf("changed configuration value. configurationKey: %v -> configuration new & old value: %v --- %v, ", config.Key, config.Value, cDto.value )
					}

					cDto.value = config.Value
					configurationDtoList[config.Key] = cDto
				}
			} else {
				log.Printf("new configuration found. configurationKey: %v configuration value: %v", config.Key, config.Value)
				configurationDtoList[config.Key] = ConfigurationDto{
					key:   config.Key,
					value: config.Value,
				}
			}
		}

		if f {
			Complete <- true
			init = true
			close(Complete)
			f = false

			if s.IntervalTimeInSecond < time.Second * 1 {
				s.IntervalTimeInSecond = time.Second * 60
				log.Printf("wrong intervalTimeInSecond value, will be set as: %s", s.IntervalTimeInSecond)
			}
		}

		time.Sleep(s.IntervalTimeInSecond)
		interval <- true
		continue
	}
}

func Init(settings Settings) {
	interval <- true
	go settings.loadConfigurationsFromService()
}