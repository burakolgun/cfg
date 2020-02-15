package cfg

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/beevik/guid"
)

type Settings struct {
	Host  string
	ProjectName          string
	IntervalTimeInSecond time.Duration
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
	Timeout: time.Second * 60,
}

var configurationDtoList = make(map[string]configurationDto)
var interval = make(chan bool, 1)

func Get(configurationKey string) configurationDto {
	return configurationDtoList[configurationKey]
}

func getConfigurationsFromService(settings Settings) ([]configuration, error) {

	req, err := http.NewRequest("GET", settings.Host+"/configurations?projectName="+settings.ProjectName, nil)

	if err != nil {
		log.Println("an error occurred while getting configurations.", err)
	}

	req.Header.Add("x-correlationId", guid.New().String())
	req.Header.Add("x-agentName", "cfg-go-client")

	resp, err := client.Do(req)
	var configurationList []configuration
	err = json.NewDecoder(resp.Body).Decode(&configurationList)

	if err != nil {
		log.Println("an error occurred while getting configurations...", err)
		return configurationList, err
	}

	defer resp.Body.Close()

	return configurationList, nil
}

func loadConfigurationsFromService(settings Settings) error {
	for {
		<-interval
		configurationList, err := getConfigurationsFromService(settings)

		if err != nil {
			log.Println("configurations didn't updated")
			return err
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

		time.Sleep(settings.IntervalTimeInSecond * time.Second)
		interval <- true
	}
}

func Init(settings Settings) error {
	if len(settings.Host) == 0 || len(settings.ProjectName) == 0 {
		log.Println("Settings must be valid")
		return errors.New("settings must be valid")
	}

	interval <- true
	go loadConfigurationsFromService(settings)

	return nil
}
