package dataSources

import (
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/identity"
)

func NewObjectStorage() (*identity.IdentityClient, error) {
	config := common.DefaultConfigProvider()
	client, err := identity.NewIdentityClientWithConfigurationProvider(config)
	if err != nil {
		return nil, err
	}

	return &client, nil
}
