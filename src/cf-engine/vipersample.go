package main

import (
	"fmt"
	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigName("sample")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic (fmt.Errorf("Fatal error config file: %s \n", err))
	}
	fmt.Println(viper.Get("RESOURCE_GROUP"))
	accounts := viper.Get("STORAGE_ACCOUNTS")
	for _,account := range accounts.([]interface{}) {
		accountProperty := account.(map[interface{}]interface{})
		fmt.Println(accountProperty["name"])
		containers := accountProperty["containers"].([]interface{})
		for _,container := range containers {
			containerProperty := container.(map[interface{}]interface{})
			fmt.Println(containerProperty["name"])
			fmt.Println(containerProperty["args"])
		}
	} 


}