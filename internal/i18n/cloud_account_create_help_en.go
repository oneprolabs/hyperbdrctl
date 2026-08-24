package i18n

const (
	cloudAccountCreateAliyunBlockEN = `Create an Alibaba Cloud block-storage cloud account.

Parameter Sources:
  --access-key-id string
    Use the Access Key ID from the Alibaba Cloud account.

  --access-key-secret string
    Use the Access Key Secret from the Alibaba Cloud account.

  --auth-region-id string
    Use the Alibaba Cloud authentication region ID.

Resource Retrieval:
  Query the available authentication regions first:
    hyperbdrctl cloud-resource fetch --cloud-type aliyun --storage-type block \
      --access-key-id <ak> \
      --access-key-secret <sk> \
      --fetch-res regions \
      --output json

The minimum create command is:
  hyperbdrctl cloud-account create --cloud-type aliyun --storage-type block \
    --access-key-id <ak> \
    --access-key-secret <sk> \
    --auth-region-id <auth_region_id>

After creation succeeds, wait for the task:
  hyperbdrctl cloud-account wait --id <account_id>

After the task finishes, view the cloud account details:
  hyperbdrctl cloud-account detail --id <account_id>
  hyperbdrctl cloud-account list --storage-type block`
	cloudAccountCreateHuaweiBlockEN = `Create a Huawei Cloud block-storage cloud account.

Parameter Sources:
  --access-key-id string
    Use the Access Key ID from the Huawei Cloud account.

  --access-key-secret string
    Use the Access Key Secret from the Huawei Cloud account.

  --auth-region-id string
    Confirm the authentication region ID before creating the cloud account.

Resource Retrieval:
  Query the region list first:
    hyperbdrctl cloud-resource fetch --cloud-type huawei --storage-type block \
      --access-key-id <ak> \
      --access-key-secret <sk> \
      --fetch-res regions \
      --output json

The minimum create command is:
  hyperbdrctl cloud-account create --cloud-type huawei --storage-type block \
    --access-key-id <ak> \
    --access-key-secret <sk> \
    --auth-region-id <auth_region_id>

To save a cloud account display name, optionally add:
  --account-name <name>

After creation succeeds, wait for the task:
  hyperbdrctl cloud-account wait --id <account_id>

After the task finishes, view the cloud account details:
  hyperbdrctl cloud-account detail --id <account_id>`
	cloudAccountCreateOpenStackBlockEN = `Create an OpenStack block-storage cloud account.

Parameter Sources:
  --auth-url string
    After logging in to the platform, click the username in the top-right corner, open OpenStack RC File, and read OS_AUTH_URL.

  --username string
    Usually use the username shown in the top-right corner of the platform.

  --password string
    Enter the login password for that username.

  --user-domain-id string
    On the OpenStack control node, run ` + cloudAccountCreateHelpBacktick + `openstack user show <username>` + cloudAccountCreateHelpBacktick + ` and read ` + cloudAccountCreateHelpBacktick + `domain_id` + cloudAccountCreateHelpBacktick + `; the default is usually ` + cloudAccountCreateHelpBacktick + `default` + cloudAccountCreateHelpBacktick + `.

  --project-domain-id string
    Read OS_PROJECT_DOMAIN_ID from OpenStack RC File; the default is usually ` + cloudAccountCreateHelpBacktick + `default` + cloudAccountCreateHelpBacktick + `.

  --project-name string
    Usually the same as the username.

  --region-name string
    Read OS_REGION_NAME from OpenStack RC File; the default is usually ` + cloudAccountCreateHelpBacktick + `RegionOne` + cloudAccountCreateHelpBacktick + `.

Resource Retrieval:
  Creating an OpenStack block-storage cloud account does not require a resource query first.
  Prepare the authentication and project flags above, then create the cloud account directly.

The minimum create command is:
  hyperbdrctl cloud-account create --cloud-type openstack --storage-type block \
    --auth-url <auth_url> \
    --username <username> \
    --password <password> \
    --user-domain-id <user_domain_id> \
    --project-domain-id <project_domain_id> \
    --project-name <project_name> \
    --region-name <region_name>

After creation succeeds, wait for the task:
  hyperbdrctl cloud-account wait --id <account_id>

After the task finishes, view the cloud account details:
  hyperbdrctl cloud-account detail --id <account_id>`
	cloudAccountCreateAliyunObjectEN = `Create an Alibaba Cloud object-storage cloud account.

Parameter Sources:
  --access-key-id string
    Use the Access Key ID from the Alibaba Cloud account.

  --access-key-secret string
    Use the Access Key Secret from the Alibaba Cloud account.

  --region-id string
    Use the target region ID, usually obtained from the region list query.

  --region-name string
    When omitted, the CLI resolves and fills the region display name from ` + cloudAccountCreateHelpBacktick + `region-id` + cloudAccountCreateHelpBacktick + `.

  --use-internal-ip string
    0 selects public access; 1 selects internal access.

Resource Retrieval:
  Query the region list first:
    hyperbdrctl cloud-resource fetch --cloud-type aliyun --storage-type object \
      --access-key-id <ak> \
      --access-key-secret <sk> \
      --fetch-res regions \
      --output json

  Query image resources only when you need to override image-related fields manually:
    hyperbdrctl cloud-resource fetch --cloud-type aliyun --storage-type object \
      --access-key-id <ak> \
      --access-key-secret <sk> \
      --region-id <region_id> \
      --fetch-res boot_loader_images,images \
      --output json

The minimum create command is:
  hyperbdrctl cloud-account create --cloud-type aliyun --storage-type object \
    --access-key-id <ak> \
    --access-key-secret <sk> \
    --region-id <region_id>

After creation succeeds, wait for the task:
  hyperbdrctl cloud-account wait --id <account_id>

After the task finishes, view the cloud account details:
  hyperbdrctl cloud-account detail --id <account_id>`
	cloudAccountCreateHuaweiObjectEN = `Create a Huawei Cloud object-storage cloud account.

Parameter Sources:
  --access-key-id string
    Use the Access Key ID from the Huawei Cloud account.

  --access-key-secret string
    Use the Access Key Secret from the Huawei Cloud account.

  --region-id string
    Confirm the target region ID before creating the cloud account.

  --custom-name string
    Add this only when you want to save a custom name.

Resource Retrieval:
  Query the region list first:
    hyperbdrctl cloud-resource fetch --cloud-type huawei --storage-type object \
      --access-key-id <ak> \
      --access-key-secret <sk> \
      --fetch-res regions \
      --output json

  Then query availability zones:
    hyperbdrctl cloud-resource fetch --cloud-type huawei --storage-type object \
      --access-key-id <ak> \
      --access-key-secret <sk> \
      --region-id <region_id> \
      --fetch-res zones \
      --output json

  Query flavors as needed:
    hyperbdrctl cloud-resource fetch --cloud-type huawei --storage-type object \
      --access-key-id <ak> \
      --access-key-secret <sk> \
      --region-id <region_id> \
      --zone-id <zone_id> \
      --fetch-res flavors \
      --flavor-vcpus 2 \
      --flavor-ram 4 \
      --output json

  Query networks and subnets as needed:
    hyperbdrctl cloud-resource fetch --cloud-type huawei --storage-type object \
      --access-key-id <ak> \
      --access-key-secret <sk> \
      --region-id <region_id> \
      --fetch-res networks,subnets \
      --output json

  Query images and system volume types as needed:
    hyperbdrctl cloud-resource fetch --cloud-type huawei --storage-type object \
      --access-key-id <ak> \
      --access-key-secret <sk> \
      --region-id <region_id> \
      --fetch-res images \
      --output json

    hyperbdrctl cloud-resource fetch --cloud-type huawei --storage-type object \
      --access-key-id <ak> \
      --access-key-secret <sk> \
      --region-id <region_id> \
      --zone-id <zone_id> \
      --fetch-res system_volume_types \
      --output json

The minimum create command is:
  hyperbdrctl cloud-account create --cloud-type huawei --storage-type object \
    --access-key-id <ak> \
    --access-key-secret <sk> \
    --region-id <region_id>

After creation succeeds, wait for the task:
  hyperbdrctl cloud-account wait --id <account_id>

After the task finishes, view the cloud account details:
  hyperbdrctl cloud-account detail --id <account_id>`
	cloudAccountCreateOpenStackObjectEN = `Create an OpenStack object-storage cloud account.

Parameter Sources:
  --auth-url string
    After logging in to the platform, click the username in the top-right corner, open OpenStack RC File, and read OS_AUTH_URL.

  --username string
    Usually use the username shown in the top-right corner of the platform.

  --password string
    Enter the login password for that username.

  --user-domain-id string
    On the OpenStack control node, run ` + cloudAccountCreateHelpBacktick + `openstack user show <username>` + cloudAccountCreateHelpBacktick + ` and read ` + cloudAccountCreateHelpBacktick + `domain_id` + cloudAccountCreateHelpBacktick + `; the default is usually ` + cloudAccountCreateHelpBacktick + `default` + cloudAccountCreateHelpBacktick + `.

  --project-domain-id string
  --project-id string
  --project-name string
  --region-id string
    These project and region fields usually come from resource retrieval results; pass them explicitly only to override automatically supplied values.

  --use-internal-ip string
    0 selects public access; 1 selects internal access.

Resource Retrieval:
  Query cloud resources first:
    hyperbdrctl cloud-resource fetch --cloud-type openstack --storage-type object \
      --auth-url <auth_url> \
      --username <username> \
      --password <password> \
      --user-domain-id <user_domain_id> \
      --output json

  Pay particular attention to these fields in the response:
    project_domain_id
    project_id
    project_name
    region_id
    region_name
    boot_loader_image_id
    boot_loader_image_name
    disk_bus_type_id
    disk_bus_type_name

The minimum create command is:
  hyperbdrctl cloud-account create --cloud-type openstack --storage-type object \
    --auth-url <auth_url> \
    --username <username> \
    --password <password> \
    --user-domain-id <user_domain_id>

After creation succeeds, wait for the task:
  hyperbdrctl cloud-account wait --id <account_id>

After the task finishes, view the cloud account details:
  hyperbdrctl cloud-account detail --id <account_id>`
)
