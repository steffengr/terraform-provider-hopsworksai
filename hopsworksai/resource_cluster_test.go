package hopsworksai

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/api"
)

func TestAccClusterAWS_basic(t *testing.T) {
	testSkipAWS(t)
	testAccCluster_basic(t, api.AWS)
}

func TestAccClusterAZURE_basic(t *testing.T) {
	testSkipAZURE(t)
	testAccCluster_basic(t, api.AZURE)
}

func testAccCluster_basic(t *testing.T, cloud api.CloudProvider) {
	resourceName := "hopsworksai_cluster.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     testAccPreCheck(t),
		Providers:    testAccProviders,
		CheckDestroy: testAccClusterCheckDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfigBasic(cloud),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "url"),
					resource.TestCheckResourceAttr(resourceName, "state", api.Running.String()),
					resource.TestCheckResourceAttr(resourceName, "activation_state", api.Stoppable.String()),
					resource.TestCheckResourceAttr(resourceName, "update_state", "none"),
					resource.TestCheckResourceAttr(resourceName, "workers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "open_ports.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "open_ports.0.ssh", "false"),
					resource.TestCheckResourceAttr(resourceName, "open_ports.0.kafka", "false"),
					resource.TestCheckResourceAttr(resourceName, "open_ports.0.feature_store", "false"),
					resource.TestCheckResourceAttr(resourceName, "open_ports.0.online_feature_store", "false"),

					resource.TestCheckResourceAttr(resourceName, strings.ToLower(cloud.String())+"_attributes.0.network.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:             testAccClusterConfig(cloud, `update_state = "start"`),
				ExpectError:        regexp.MustCompile("cluster is already running"),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccClusterConfig(cloud, fmt.Sprintf(`
				workers{
					instance_type = "%s"
					disk_size = 256
					count = 2
				}`, testWorkerInstanceType1(cloud))),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "state", api.Running.String()),
					resource.TestCheckResourceAttr(resourceName, "activation_state", api.Stoppable.String()),
					resource.TestCheckResourceAttr(resourceName, "update_state", "none"),
					resource.TestCheckResourceAttr(resourceName, "workers.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workers.*", map[string]string{
						"instance_type": testWorkerInstanceType1(cloud),
						"disk_size":     "256",
						"count":         "2",
					}),
				),
			},
			{
				Config: testAccClusterConfig(cloud, fmt.Sprintf(`
				workers{
					instance_type = "%s"
					disk_size = 256
					count = 1
				}
				workers{
					instance_type = "%s"
					disk_size = 512
					count = 1
				}
				`, testWorkerInstanceType1(cloud), testWorkerInstanceType1(cloud))),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "state", api.Running.String()),
					resource.TestCheckResourceAttr(resourceName, "activation_state", api.Stoppable.String()),
					resource.TestCheckResourceAttr(resourceName, "update_state", "none"),
					resource.TestCheckResourceAttr(resourceName, "workers.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workers.*", map[string]string{
						"instance_type": testWorkerInstanceType1(cloud),
						"disk_size":     "256",
						"count":         "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workers.*", map[string]string{
						"instance_type": testWorkerInstanceType1(cloud),
						"disk_size":     "512",
						"count":         "1",
					}),
				),
			},
			{
				Config: testAccClusterConfig(cloud, fmt.Sprintf(`
				workers{
					instance_type = "%s"
					disk_size = 256
					count = 1
				}
				workers{
					instance_type = "%s"
					disk_size = 512
					count = 1
				}
				workers{
					instance_type = "%s"
					disk_size = 512
					count = 1
				}
				`, testWorkerInstanceType1(cloud), testWorkerInstanceType1(cloud), testWorkerInstanceType2(cloud))),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "state", api.Running.String()),
					resource.TestCheckResourceAttr(resourceName, "activation_state", api.Stoppable.String()),
					resource.TestCheckResourceAttr(resourceName, "update_state", "none"),
					resource.TestCheckResourceAttr(resourceName, "workers.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workers.*", map[string]string{
						"instance_type": testWorkerInstanceType1(cloud),
						"disk_size":     "256",
						"count":         "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workers.*", map[string]string{
						"instance_type": testWorkerInstanceType1(cloud),
						"disk_size":     "512",
						"count":         "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workers.*", map[string]string{
						"instance_type": testWorkerInstanceType2(cloud),
						"disk_size":     "512",
						"count":         "1",
					}),
				),
			},
			{
				Config: testAccClusterConfig(cloud, fmt.Sprintf(`
				workers{
					instance_type = "%s"
					disk_size = 512
					count = 1
				}
				`, testWorkerInstanceType2(cloud))),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "state", api.Running.String()),
					resource.TestCheckResourceAttr(resourceName, "activation_state", api.Stoppable.String()),
					resource.TestCheckResourceAttr(resourceName, "update_state", "none"),
					resource.TestCheckResourceAttr(resourceName, "workers.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workers.*", map[string]string{
						"instance_type": testWorkerInstanceType2(cloud),
						"disk_size":     "512",
						"count":         "1",
					}),
				),
			},
			{
				Config: testAccClusterConfig(cloud, fmt.Sprintf(`
				workers{
					instance_type = "%s"
					disk_size = 512
					count = 2
				}
				workers{
					instance_type = "%s"
					disk_size = 512
					count = 1
				}
				`, testWorkerInstanceType2(cloud), testWorkerInstanceType1(cloud))),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "state", api.Running.String()),
					resource.TestCheckResourceAttr(resourceName, "activation_state", api.Stoppable.String()),
					resource.TestCheckResourceAttr(resourceName, "update_state", "none"),
					resource.TestCheckResourceAttr(resourceName, "workers.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workers.*", map[string]string{
						"instance_type": testWorkerInstanceType2(cloud),
						"disk_size":     "512",
						"count":         "2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workers.*", map[string]string{
						"instance_type": testWorkerInstanceType1(cloud),
						"disk_size":     "512",
						"count":         "1",
					}),
				),
			},
			{
				Config: testAccClusterConfig(cloud, fmt.Sprintf(`
				workers{
					instance_type = "%s"
					disk_size = 512
					count = 1
				}
				`, testWorkerInstanceType2(cloud))),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "state", api.Running.String()),
					resource.TestCheckResourceAttr(resourceName, "activation_state", api.Stoppable.String()),
					resource.TestCheckResourceAttr(resourceName, "update_state", "none"),
					resource.TestCheckResourceAttr(resourceName, "workers.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workers.*", map[string]string{
						"instance_type": testWorkerInstanceType2(cloud),
						"disk_size":     "512",
						"count":         "1",
					}),
				),
			},
			{
				Config: testAccClusterConfig(cloud, `
				open_ports{
					ssh = true
					kafka = true
				}`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "open_ports.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "open_ports.0.ssh", "true"),
					resource.TestCheckResourceAttr(resourceName, "open_ports.0.kafka", "true"),
					resource.TestCheckResourceAttr(resourceName, "open_ports.0.feature_store", "false"),
					resource.TestCheckResourceAttr(resourceName, "open_ports.0.online_feature_store", "false"),
				),
			},
			{
				Config: testAccClusterConfig(cloud, `
				open_ports{
					feature_store = true
					online_feature_store = true
				}`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "open_ports.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "open_ports.0.ssh", "false"),
					resource.TestCheckResourceAttr(resourceName, "open_ports.0.kafka", "false"),
					resource.TestCheckResourceAttr(resourceName, "open_ports.0.feature_store", "true"),
					resource.TestCheckResourceAttr(resourceName, "open_ports.0.online_feature_store", "true"),
				),
			},
			{
				Config: testAccClusterConfig(cloud, `update_state = "stop"`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "state", api.Stopped.String()),
					resource.TestCheckResourceAttr(resourceName, "activation_state", api.Startable.String()),
					resource.TestCheckResourceAttr(resourceName, "update_state", "stop"),
				),
			},
			{
				Config:             testAccClusterConfig(cloud, `update_state = "stop"`),
				ExpectError:        regexp.MustCompile("cluster is already stopped"),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testWorkerInstanceType1(cloud api.CloudProvider) string {
	return testWorkerInstanceType(cloud, true)
}

func testWorkerInstanceType2(cloud api.CloudProvider) string {
	return testWorkerInstanceType(cloud, false)
}

func testWorkerInstanceType(cloud api.CloudProvider, alternative bool) string {
	if cloud == api.AWS {
		if alternative {
			return "t3a.medium"
		} else {
			return "t3a.large"
		}
	} else if cloud == api.AZURE {
		if alternative {
			return "Standard_D4_v3"
		} else {
			return "Standard_D8_v3"
		}
	}
	return ""
}

func testAccClusterCheckDestroy() func(s *terraform.State) error {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*api.HopsworksAIClient)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "hopsworksai_cluster" {
				continue
			}
			cluster, err := api.GetCluster(context.Background(), client, rs.Primary.ID)
			if err != nil {
				return err
			}

			if cluster != nil {
				return fmt.Errorf("found unterminated cluster %s", rs.Primary.ID)
			}
		}
		return nil
	}
}

func testAccClusterConfigBasic(cloud api.CloudProvider) string {
	return testAccClusterConfig(cloud, "")
}

func testAccClusterConfig(cloud api.CloudProvider, extraConfig string) string {
	return fmt.Sprintf(`
	resource "hopsworksai_cluster" "test" {
		name    = "%sr%s"
		ssh_key = "%s"	  
		head {
		}
		
		%s
		
		%s 

		tags = {
		  "Purpose" = "acceptance-test"
		}
	  }
	`, clusterPrefixName, strings.ToLower(cloud.String()), testAccClusterCloudSSHKeyAttribute(cloud), testAccClusterCloudConfigAttributes(cloud), extraConfig)
}

func testAccClusterCloudSSHKeyAttribute(cloud api.CloudProvider) string {
	if cloud == api.AWS {
		return os.Getenv(env_AWS_SSH_KEY)
	} else if cloud == api.AZURE {
		return os.Getenv(env_AZURE_SSH_KEY)
	}
	return ""
}

func testAccClusterCloudConfigAttributes(cloud api.CloudProvider) string {
	if cloud == api.AWS {
		return fmt.Sprintf(`
		aws_attributes {
			region               = "%s"
			instance_profile_arn = "%s"
			bucket_name          = "%s"
		  }
		`, os.Getenv(env_AWS_REGION), os.Getenv(env_AWS_INSTANCE_PROFILE_ARN), os.Getenv(env_AWS_BUCKET_NAME))
	} else if cloud == api.AZURE {
		return fmt.Sprintf(`
		azure_attributes {
			location                       = "%s"
			resource_group                 = "%s"
			storage_account                = "%s"
			user_assigned_managed_identity = "%s"
		  }
		`, os.Getenv(env_AZURE_LOCATION), os.Getenv(env_AZURE_RESOURCE_GROUP), os.Getenv(env_AZURE_STORAGE_ACCOUNT), os.Getenv(env_AZURE_USER_ASSIGNED_IDENTITY_NAME))
	}
	return ""
}
