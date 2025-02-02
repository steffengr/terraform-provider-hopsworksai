---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "Hopsworks.ai Provider"
subcategory: ""
description: |-
  The Hopsworks.ai provider is used to interact with [Hopsworks.ai](https://managed.hopsworks.ai)
---

# Hopsworks.ai Provider

The Hopsworksai terraform provider is used to interact with [Hopsworks.ai](https://managed.hopsworks.ai) to manage Hopsworks clusters and Hopsworks Feature Store in the cloud.
If you are new to Hopsworks, then first you need to create an account on [Hopsworks.ai](https://managed.hopsworks.ai), and then you can follow one of the getting started guides to connect either your AWS account or Azure account to create your own Hopsworks clusters. 
  * [Getting Started with AWS](https://docs.hopsworks.ai/latest/setup_installation/aws/getting_started/)
  * [Getting Started with Azure](https://docs.hopsworks.ai/latest/setup_installation/azure/getting_started/)


-> A Hopsworks API Key is required to allow the provider to manage clusters on Hopsworks.ai on your behalf. To create an API Key, follow [this guide](https://docs.hopsworks.ai/latest/setup_installation/common/api_key).

In the following sections, we show two usage examples to create Hopsworks clusters on AWS and Azure, for more detailed examples check the [examples/complete](https://github.com/logicalclocks/terraform-provider-hopsworksai/tree/main/examples/complete) directory in the git repository.

## AWS Example Usage 

Hopsworks.ai deploys Hopsworks clusters to your AWS account using the permissions provided during [account setup](https://docs.hopsworks.ai/latest/setup_installation/aws/getting_started/#step-1-connecting-your-aws-account). 
To create a Hopsworks cluster, you will need to create an empty S3 bucket, an ssh key, and an instance profile with the required [Hopsworks permissions](https://docs.hopsworks.ai/latest/setup_installation/aws/getting_started/#step-2-creating-instance-profile). 
If you have already created these 3 resources, you can skip the first step in the following terraform example and instead fill the corresponding attributes in Step 2 (*bucket_name*, *ssh_key*, *instance_profile_arn*) with your configuration.
Otherwise, you need to setup the credentials for your AWS account locally as described [here](https://registry.terraform.io/providers/hashicorp/aws/latest/docs), then you can run the following terraform example which creates the required AWS resources and a Hopsworks cluster. 

```terraform
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "4.16.0"
    }
    hopsworksai = {
      source = "logicalclocks/hopsworksai"
    }
  }
}

variable "region" {
  type    = string
  default = "us-east-2"
}

provider "aws" {
  region = var.region
}

provider "hopsworksai" {
  # Highly recommended to use the HOPSWORKSAI_API_KEY environment variable instead
  api_key = "YOUR HOPSWORKS API KEY"
}

# Step 1: create the required aws resources, an ssh key, an s3 bucket, and an instance profile with the required hopsworks permissions
module "aws" {
  source  = "logicalclocks/helpers/hopsworksai//modules/aws"
  region  = var.region
  version = "2.3.0"
}

# Step 2: create a cluster with 1 worker

data "hopsworksai_instance_type" "head" {
  cloud_provider = "AWS"
  node_type      = "head"
  region         = var.region
}

data "hopsworksai_instance_type" "rondb_data" {
  cloud_provider = "AWS"
  node_type      = "rondb_data"
  region         = var.region
}

data "hopsworksai_instance_type" "smallest_worker" {
  cloud_provider = "AWS"
  node_type      = "worker"
  region         = var.region
}

resource "hopsworksai_cluster" "cluster" {
  name    = "tf-hopsworks-cluster"
  ssh_key = module.aws.ssh_key_pair_name

  head {
    instance_type = data.hopsworksai_instance_type.head.id
  }

  workers {
    instance_type = data.hopsworksai_instance_type.smallest_worker.id
    count         = 1
  }

  aws_attributes {
    region               = var.region
    instance_profile_arn = module.aws.instance_profile_arn
    bucket {
      name = module.aws.bucket_name
    }
  }

  rondb {
    single_node {
      instance_type = data.hopsworksai_instance_type.rondb_data.id
    }
  }

  open_ports {
    ssh = true
  }
}

# Outputs the url of the newly created cluster 
output "hopsworks_cluster_url" {
  value = hopsworksai_cluster.cluster.url
}
```

## Azure Example Usage 

Similar to AWS, Hopsworks.ai deploys Hopsworks clusters to your Azure account using the permissions provided during [account setup](https://docs.hopsworks.ai/latest/setup_installation/azure/getting_started/#step-1-connecting-your-azure-account). 
To create a Hopsworks cluster, you will need to create a storage account, an ssh key, and a user assigned managed identity with the required [Hopsworks permissions](https://docs.hopsworks.ai/latest/setup_installation/azure/getting_started/#step-21-creating-a-restrictive-role-for-accessing-storage)
If you have already created these 3 resources, you can skip the first step in the following terraform example and instead fill the corresponding attributes in Step 2 (*storage_account*, *ssh_key*, *user_assigned_managed_identity*) with your configuration.
Otherwise, you need to setup the credentials for your Azure account locally as described [here](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs), then you can run the following terraform example which creates the required Azure resources and a Hopsworks cluster. 
Notice that you need to replace "*YOUR AZURE RESOURCE GROUP*" with the resource group that you want to use for this cluster.


```terraform
terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "3.8.0"
    }
    hopsworksai = {
      source = "logicalclocks/hopsworksai"
    }
  }
}

variable "resource_group" {
  type    = string
  default = "YOUR AZURE RESOURCE GROUP"
}

provider "azurerm" {
  features {}
  skip_provider_registration = true
}

provider "hopsworksai" {
  # Highly recommended to use the HOPSWORKSAI_API_KEY environment variable instead
  api_key = "YOUR HOPSWORKS API KEY"
}

data "azurerm_resource_group" "rg" {
  name = var.resource_group
}

# Step 1: create the required azure resources, an ssh key, a storage account, and an user assigned managed identity with the required hopsworks permissions
module "azure" {
  source         = "logicalclocks/helpers/hopsworksai//modules/azure"
  resource_group = var.resource_group
  version        = "2.3.0"
}

# Step 2: create a cluster with no workers

data "hopsworksai_instance_type" "head" {
  cloud_provider = "AZURE"
  node_type      = "head"
  region         = module.azure.location
}

data "hopsworksai_instance_type" "rondb_data" {
  cloud_provider = "AZURE"
  node_type      = "rondb_data"
  region         = module.azure.location
}

data "hopsworksai_instance_type" "smallest_worker" {
  cloud_provider = "AZURE"
  node_type      = "worker"
  region         = module.azure.location
}

resource "azurerm_container_registry" "acr" {
  name                = "tfhopsworksbasic"
  resource_group_name = module.azure.resource_group
  location            = module.azure.location
  sku                 = "Premium"
  admin_enabled       = false
  retention_policy {
    enabled = true
    days    = 7
  }
}

resource "hopsworksai_cluster" "cluster" {
  name    = "tf-hopsworks-cluster"
  ssh_key = module.azure.ssh_key_pair_name

  head {
    instance_type = data.hopsworksai_instance_type.head.id
  }

  workers {
    instance_type = data.hopsworksai_instance_type.smallest_worker.id
    count         = 1
  }

  azure_attributes {
    location                       = module.azure.location
    resource_group                 = module.azure.resource_group
    user_assigned_managed_identity = module.azure.user_assigned_identity_name
    container {
      storage_account = module.azure.storage_account_name
    }
    acr_registry_name = azurerm_container_registry.acr.name
  }

  rondb {
    single_node {
      instance_type = data.hopsworksai_instance_type.rondb_data.id
    }
  }

  open_ports {
    ssh = true
  }
}

# Outputs the url of the newly created cluster 
output "hopsworks_cluster_url" {
  value = hopsworksai_cluster.cluster.url
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `api_gateway` (String) URL of the API Gateway to use. It is intended for development purposes only. Defaults to `https://api.hopsworks.ai`.
- `api_key` (String, Sensitive) The API Key to use to connect to your account on Hopsworka.ai. Can be specified using the HOPSWORKSAI_API_KEY environment variable.