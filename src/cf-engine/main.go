package main

import (
	"fmt"
	"log"
	"os"
//	"flag"
//	"io"
//    "gopkg.in/yaml.v2"

	"github.com/Azure/azure-sdk-for-go/arm/examples/helpers"
	"github.com/Azure/azure-sdk-for-go/arm/storage"
	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
	storagem "github.com/Azure/azure-sdk-for-go/storage"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
)

func main() {
	resourceGroup := os.Getenv("RESOURCE_GROUP")
	name   := os.Getenv("STORAGE_ACCOUNT_NAME")

	c := map[string]string {
		"AZURE_CLIENT_ID": os.Getenv("AZURE_CLIENT_ID"),
		"AZURE_CLIENT_SECRET": os.Getenv("AZURE_CLIENT_SECRET"),
		"AZURE_SUBSCRIPTION_ID": os.Getenv("AZURE_SUBSCRIPTION_ID"),
		"AZURE_TENANT_ID": os.Getenv("AZURE_TENANT_ID")}
	if err := checkEnvVar(&c); err != nil {
		log.Fatalf("Error: %v", err)
		return
	}	


	spt, err := helpers.NewServicePrincipalTokenFromCredentials(c, azure.PublicCloud.ResourceManagerEndpoint)
	if err != nil {
		log.Fatalf("Error: %v", err)
		return 
	}

	gp := resources.Group {
		Location: to.StringPtr("japaneast")}
	gc := resources.NewGroupsClient(c["AZURE_SUBSCRIPTION_ID"])
	gc.Authorizer = spt
	_, err = gc.CreateOrUpdate(resourceGroup, gp)
	if err != nil {
		log.Fatalf("Error: %v", err)
		return 
	}

	
	ac := storage.NewAccountsClient(c["AZURE_SUBSCRIPTION_ID"])
	ac.Authorizer = spt

	cna, err := ac.CheckNameAvailability(
		storage.AccountCheckNameAvailabilityParameters {
			Name: to.StringPtr(name),
			Type: to.StringPtr("Microsoft.Storage/storageAccounts")})

	if err != nil {
		log.Fatalf("Error: %v", err)
		return
	}

	if !to.Bool(cna.NameAvailable) {
		fmt.Printf("%s is unavailable -- try with another name\n", name)
		return
	}
	fmt.Printf("%s is available\n\n", name)

	cp := storage.AccountCreateParameters{
		Sku: &storage.Sku{
			Name: storage.StandardLRS,
			Tier: storage.Standard},
			Location: to.StringPtr("japaneast")}
	cancel := make(chan struct{})
	if _, err = ac.Create(resourceGroup, name, cp, cancel); err != nil {
		fmt.Printf("Create '%s' storage account failed: %v\n", name, err)

	}

	keyResults, err := ac.ListKeys(resourceGroup, name)
	if err != nil {
		log.Fatalf("Error: %v", err)
		return
	}
	accountKeyList := keyResults.Keys
	pl := *accountKeyList
	accountKey := pl[0]
	value   := accountKey.Value
	fmt.Printf("AccountKey: %s\nValue: %s", name, *value)

	createContainer(name, *value, os.Getenv("CONTAINER_NAME"))


}

func createContainer(storageAccountName string, storageAccountKeyValue string, containerName string) {
	client, err := storagem.NewBasicClient(storageAccountName, storageAccountKeyValue)
	if err != nil {
		log.Fatalf("Error: %v", err)
		return
	}

	blobClient := client.GetBlobService()
	cnt := blobClient.GetContainerReference(containerName)
	_, err = cnt.CreateIfNotExists()
	if err != nil {
		log.Fatalf("Error: %v", err)
		return
	}
	fmt.Printf("Container %s has been created", containerName)
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