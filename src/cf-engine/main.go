package main

import (
	"errors"
	"flag"
	"fmt"
	"log"

	"github.com/spf13/viper"

	"github.com/Azure/azure-sdk-for-go/arm/examples/helpers"
	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
	"github.com/Azure/azure-sdk-for-go/arm/storage"
	storagem "github.com/Azure/azure-sdk-for-go/storage"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
)

func main() {
	fileName := flag.String("f", "config", "Configuration File: Path to Configuration file")
	flag.Parse()
	viper.SetConfigName(*fileName)
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error: config file %s", *fileName))
	}
	resourceGroup := viper.Get("RESOURCE_GROUP").(string)
	storageAccounts := viper.Get("STORAGE_ACCOUNTS")

	strageConfig := map[string]string{
		"AZURE_CLIENT_ID":       viper.Get("AZURE_CLIENT_ID").(string),
		"AZURE_CLIENT_SECRET":   viper.Get("AZURE_CLIENT_SECRET").(string),
		"AZURE_SUBSCRIPTION_ID": viper.Get("AZURE_SUBSCRIPTION_ID").(string),
		"AZURE_TENANT_ID":       viper.Get("AZURE_TENANT_ID").(string)}
	if err := checkEnvVar(&strageConfig); err != nil {
		log.Println("Error: %v", err)
		return
	}

	spt, err := helpers.NewServicePrincipalTokenFromCredentials(strageConfig, azure.PublicCloud.ResourceManagerEndpoint)
	if err != nil {
		log.Println("Error: %v", err)
		return
	}

	group := resources.Group{Location: to.StringPtr("japaneast")}
	groupsClient := resources.NewGroupsClient(strageConfig["AZURE_SUBSCRIPTION_ID"])
	groupsClient.Authorizer = spt
	_, err = groupsClient.CreateOrUpdate(resourceGroup, group)
	if err != nil {
		log.Println("Error: %v", err)
		return
	}

	accoutClient := storage.NewAccountsClient(strageConfig["AZURE_SUBSCRIPTION_ID"])
	accoutClient.Authorizer = spt

	for _, account := range storageAccounts.([]interface{}) {
		accountProperty := account.(map[interface{}]interface{})
		accountKey, err := createStrageAccount(accoutClient, resourceGroup, accountProperty["name"].(string))
		if err != nil {
			log.Printf("Storage Account Creation Error : %v", err)
			continue
		}
		containers := accountProperty["containers"].([]interface{})
		for _, container := range containers {
			containerProperty := container.(map[interface{}]interface{})
			createContainer(accountProperty["name"].(string), accountKey, containerProperty["name"].(string))
			if err != nil {
				log.Printf("Container Creation Error : %v", err)
				continue
			}
		}
	}
}

func createContainer(storageAccountName string, storageAccountKeyValue string, containerName string) error {
	client, err := storagem.NewBasicClient(storageAccountName, storageAccountKeyValue)
	if err != nil {
		log.Printf("Error: %v", err)
		return err
	}

	blobClient := client.GetBlobService()
	container := blobClient.GetContainerReference(containerName)
	_, err = container.CreateIfNotExists()
	if err != nil {
		log.Println("Error: %v", err)
		return err
	}
	log.Printf("Container %s has been created\n", containerName)
	return nil
}

func checkEnvVar(envVars *map[string]string) error {
	var missingVars []string
	for varName, value := range *envVars {
		if value == "" {
			missingVars = append(missingVars, varName)
		}
	}
	if len(missingVars) > 0 {
		return fmt.Errorf("Missing environment variables %v", missingVars)
	}
	return nil
}

func createStrageAccount(accoutClient storage.AccountsClient, resourceGroup string, strageName string) (string, error) {
	cna, err := accoutClient.CheckNameAvailability(
		storage.AccountCheckNameAvailabilityParameters{
			Name: to.StringPtr(strageName),
			Type: to.StringPtr("Microsoft.Storage/storageAccounts")})
	if err != nil {
		log.Printf("Error: %v", err)
		return "", err
	}

	if !to.Bool(cna.NameAvailable) {
		return "", errors.New("Strage name \"" + strageName + "\" is unavailable")
	}

	cp := storage.AccountCreateParameters{
		Sku: &storage.Sku{
			Name: storage.StandardLRS,
			Tier: storage.Standard},
		Location: to.StringPtr("japaneast")}
	cancel := make(chan struct{})
	if _, err = accoutClient.Create(resourceGroup, strageName, cp, cancel); err != nil {
		log.Println("Create '%s' storage account failed: %v\n", strageName, err)
	}

	keyResults, err := accoutClient.ListKeys(resourceGroup, strageName)
	if err != nil {
		log.Println("Error: %v", err)
		return "", err
	}
	accountKeyList := keyResults.Keys
	pl := *accountKeyList
	accountKey := pl[0]
	value := accountKey.Value
	log.Printf("AccountKey: %s\nValue: %s\n", strageName, *value)

	return *value, nil
}
